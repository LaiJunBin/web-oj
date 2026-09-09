package executejudgment

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"judge-backend/internal/app/judge_domain/events/judged"
	"judge-backend/internal/app/judge_domain/events/judgeupdated"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/execresult"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/lifecycle"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/status"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/testcase"
	"judge-backend/internal/app/judge_domain/judge/domain/repository"
	domainservice "judge-backend/internal/app/judge_domain/judge/domain/service"
	sharedtypes "judge-backend/internal/app/shared/types"
)

// Usecase executes one judgment from start to terminal status. It is the
// application port the judge worker drives directly; the runtime behind it is
// whatever ExecutionEnvironmentProvisioner implementation was wired in.
type Usecase interface {
	Execute(ctx context.Context, logWriter io.Writer, j *judgment.Judgment, workerID string) error
}

// DefaultProvisionRetryBackoff is the pause between provisioning attempts.
// The retry exists for a loaded host (project container missing its
// HEALTHCHECK window); retrying immediately would hit the same load spike,
// while waiting costs judgment-budget time — a few seconds balances both.
const DefaultProvisionRetryBackoff = 5 * time.Second

type usecase struct {
	provisioner           domainservice.ExecutionEnvironmentProvisioner
	submissionService     domainservice.SubmissionService
	submissionRepo        repository.SubmissionRepository
	userRepo              repository.UserRepository
	judgmentRepo          repository.JudgmentRepository
	eventBus              sharedtypes.EventBus
	lifecycleTimeout      time.Duration
	testcaseTimeout       time.Duration
	provisionMaxAttempts  int
	provisionRetryBackoff time.Duration
	logger                sharedtypes.Logger
}

// NewUsecase takes the timeouts and retry policy as plain values: the
// application layer must not know where configuration comes from — the
// composition root extracts them from whatever config source it uses.
func NewUsecase(
	provisioner domainservice.ExecutionEnvironmentProvisioner,
	submissionService domainservice.SubmissionService,
	submissionRepo repository.SubmissionRepository,
	userRepo repository.UserRepository,
	judgmentRepo repository.JudgmentRepository,
	eventBus sharedtypes.EventBus,
	testcaseTimeout time.Duration,
	lifecycleTimeout time.Duration,
	provisionMaxAttempts int,
	provisionRetryBackoff time.Duration,
	logger sharedtypes.Logger,
) Usecase {
	if provisionMaxAttempts < 1 {
		provisionMaxAttempts = 1
	}
	return &usecase{
		provisioner:           provisioner,
		submissionService:     submissionService,
		submissionRepo:        submissionRepo,
		userRepo:              userRepo,
		judgmentRepo:          judgmentRepo,
		eventBus:              eventBus,
		lifecycleTimeout:      lifecycleTimeout,
		testcaseTimeout:       testcaseTimeout,
		provisionMaxAttempts:  provisionMaxAttempts,
		provisionRetryBackoff: provisionRetryBackoff,
		logger:                logger,
	}
}

// errJudgmentNoLongerJudging is returned by the progress-save helpers when
// the document's status is no longer Judging — an out-of-band cancel (or,
// in principle, anything else) claimed this judgment while it was running.
// Continuing to execute testcases and compute results for a judgment that
// isn't ours to decide anymore is wasted work at best; the eventual
// terminal save is guarded the same way regardless, so aborting here saves
// nothing by continuing.
var errJudgmentNoLongerJudging = errors.New("judgment is no longer Judging")

func (u *usecase) Execute(ctx context.Context, logWriter io.Writer, j *judgment.Judgment, workerID string) error {
	if err := j.Start(); err != nil {
		u.logger.Errorf("[ExecuteJudgment] Cannot start judgment %s: %v", j.ID(), err)
		return err
	}
	// TryStart is this worker's ownership claim, not just a save: j.Start()
	// above only validated j's own popped copy, which cannot know the
	// document was cancelled while queued, nor that another holder of the
	// same queue entry got here first. See TryStart's doc comment.
	applied, err := u.judgmentRepo.TryStart(j.ID(), *j.JudgingStartedAt())
	if err != nil {
		u.logger.Errorf("[ExecuteJudgment] Failed to start judgment %s: %v", j.ID(), err)
		return err
	}
	if !applied {
		u.logger.Infof("[ExecuteJudgment] Judgment %s was no longer Queued (likely cancelled while still queued); aborting", j.ID())
		return fmt.Errorf("judgment %s is no longer queued", j.ID())
	}
	u.logger.Infof("[ExecuteJudgment] Starting judgment %s", j.ID())
	u.eventBus.Publish(judgeupdated.NewJudgeUpdatedEvent(j))

	runErr := u.runSafely(ctx, logWriter, j, workerID)

	// Each branch below fails closed: if the domain transition is rejected,
	// j is not in the state this branch assumed, so persisting or publishing
	// whatever it happens to hold would announce an outcome nobody computed.
	// Return without touching the repository or the event bus, and leave the
	// document to the same recovery path as any other unconfirmed terminal
	// state.
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		// The worker's emergency budget expired: this is a fault (hung
		// container, stuck healthcheck), not a user cancellation.
		u.logger.Errorf("[ExecuteJudgment] Judgment %s exceeded the maximum allowed duration", j.ID())
		if err := j.Fail(); err != nil {
			u.logger.Errorf("[ExecuteJudgment] Cannot mark judgment %s as failed: %v", j.ID(), err)
			return err
		}
		j.SetErrorMessage("judgment exceeded the maximum allowed duration")
	} else if ctx.Err() != nil {
		u.logger.Infof("[ExecuteJudgment] Judgment %s was cancelled", j.ID())
		if err := j.Cancel(""); err != nil {
			u.logger.Errorf("[ExecuteJudgment] Cannot cancel judgment %s: %v", j.ID(), err)
			return err
		}
	} else if runErr != nil {
		u.logger.Errorf("[ExecuteJudgment] Judgment %s failed: %v", j.ID(), runErr)
		if err := j.Fail(); err != nil {
			u.logger.Errorf("[ExecuteJudgment] Cannot mark judgment %s as failed: %v", j.ID(), err)
			return err
		}
		var lcErr *lifecycleFailedError
		if errors.As(runErr, &lcErr) {
			// The reason is what a contestant may see; the error message is
			// the operator's, and neither may carry the command or output.
			j.SetReason(LifecycleSetupFailedReason)
			cancelUnfinishedTestcases(j)
		}
		j.SetErrorMessage(runErr.Error())
	} else {
		if err := j.Finish(); err != nil {
			u.logger.Errorf("[ExecuteJudgment] Cannot finish judgment %s: %v", j.ID(), err)
			return err
		}
	}
	published, terminalErr := u.saveTerminalAndNotify(j)
	if terminalErr != nil {
		// Deliberately does NOT publish judged below: we could not confirm
		// what the persisted state actually is (a SaveTerminalIfJudging
		// error, or losing the race and then failing to re-fetch the real
		// state too). Publishing a terminal event for an outcome that was
		// never confirmed persisted would let downstream consumers believe
		// a status the DB doesn't actually have.
		u.logger.Errorf("[ExecuteJudgment] %v", terminalErr)
		return terminalErr
	}
	u.logger.Infof("[ExecuteJudgment] Judgment %s finished with status %s", published.ID(), published.Status().String())

	u.eventBus.Publish(judged.NewJudgedEvent(published))
	return runErr
}

// abortIfNoLongerJudging translates a SaveProgressIfJudging result into the
// shared abort-the-run error contract every progress-save call site uses:
// a repository error propagates as-is, and applied=false becomes
// errJudgmentNoLongerJudging.
func abortIfNoLongerJudging(applied bool, err error) error {
	if err != nil {
		return err
	}
	if !applied {
		return errJudgmentNoLongerJudging
	}
	return nil
}

// saveAndPublish is the abort-on-failure variant used mid-run: persists
// in-progress state conditionally via SaveProgressIfJudging (see its doc
// comment) and publishes the update only once persisted. A save that
// doesn't apply aborts the judgment same as a hard failure — an out-of-band
// cancel already claimed it, so there is nothing left to gain by continuing
// to compute results for it.
func (u *usecase) saveAndPublish(j *judgment.Judgment) error {
	if err := abortIfNoLongerJudging(u.judgmentRepo.SaveProgressIfJudging(j)); err != nil {
		return err
	}
	u.eventBus.Publish(judgeupdated.NewJudgeUpdatedEvent(j))
	return nil
}

// saveAndNotify is the best-effort variant used where the caller is already
// unwinding via a different error path and this save is just best-effort
// bookkeeping (marking a testcase Cancelled after an exec failure) — it
// never aborts the caller. Still goes through SaveProgressIfJudging, not a
// plain Save: even a best-effort save must not resurrect a judgment an
// out-of-band cancel already moved out of Judging.
func (u *usecase) saveAndNotify(j *judgment.Judgment) {
	applied, err := u.judgmentRepo.SaveProgressIfJudging(j)
	if err != nil {
		u.logger.Errorf("[ExecuteJudgment] Failed to save judgment %s: %v", j.ID(), err)
		return
	}
	if !applied {
		u.logger.Debugf("[ExecuteJudgment] Judgment %s was no longer Judging, skipped progress save", j.ID())
		return
	}
	u.eventBus.Publish(judgeupdated.NewJudgeUpdatedEvent(j))
}

// saveTerminalAndNotify persists j's computed terminal outcome conditionally
// and returns the judgment whose state should actually be published — j
// itself if the save applied, otherwise the re-fetched current DB state,
// since a declined save means j's computed outcome was never stored. Returns
// an error when what got persisted could not be determined at all; callers
// must publish nothing in that case.
func (u *usecase) saveTerminalAndNotify(j *judgment.Judgment) (*judgment.Judgment, error) {
	applied, err := u.judgmentRepo.SaveTerminalIfJudging(j)
	if err != nil {
		return nil, fmt.Errorf("failed to save terminal outcome for judgment %s: %w", j.ID(), err)
	}
	if !applied {
		u.logger.Infof("[ExecuteJudgment] Judgment %s was already terminal (likely cancelled) when this worker finished; discarding this worker's computed result", j.ID())
		latest, ferr := u.judgmentRepo.FindByID(j.ID())
		if ferr != nil {
			return nil, fmt.Errorf("lost the terminal-save race for judgment %s and failed to re-fetch its actual state: %w", j.ID(), ferr)
		}
		j = latest
	}
	u.eventBus.Publish(judgeupdated.NewJudgeUpdatedEvent(j))
	return j, nil
}

// runSafely converts a panic anywhere in the run into a normal error so the
// judgment still reaches a terminal status and the worker survives.
func (u *usecase) runSafely(ctx context.Context, logWriter io.Writer, j *judgment.Judgment, workerID string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			u.logger.Errorf("[ExecuteJudgment] Recovered from panic for judgment %s: %v", j.ID(), r)
			err = fmt.Errorf("judgment execution panicked: %v", r)
		}
	}()
	return u.run(ctx, logWriter, j, workerID)
}

func (u *usecase) run(ctx context.Context, logWriter io.Writer, j *judgment.Judgment, workerID string) error {
	metadata, err := u.submissionService.LoadMetadata(j.SubmissionID())
	if err != nil {
		return err
	}

	// Resolved at execution time rather than carried on the judgment document,
	// so judgments queued before this field existed (and rejudges of them)
	// pick it up the same way.
	sub, err := u.submissionRepo.FindByID(j.SubmissionID())
	if err != nil {
		return err
	}
	submitterUsername, err := u.userRepo.FindUsernameByID(sub.UserID())
	if err != nil {
		return err
	}

	env, err := u.provision(ctx, logWriter, &domainservice.EnvironmentSpec{
		JudgmentID:        j.ID(),
		SubmissionID:      j.SubmissionID(),
		WorkerID:          workerID,
		Project:           j.Project(),
		Metadata:          metadata,
		SubmitterUsername: submitterUsername,
		DiagnosticsSink:   logWriter,
	})
	if err != nil {
		return err
	}
	defer env.Teardown()

	logCtx, stopLogStream := context.WithCancel(ctx)
	logStreamDone := make(chan struct{})
	go func() {
		defer close(logStreamDone)
		if err := env.StreamLogs(logCtx, logWriter); err != nil && logCtx.Err() == nil {
			u.logger.Errorf("[ExecuteJudgment] Error streaming logs for judgment %s: %v", j.ID(), err)
			logWriter.Write([]byte(fmt.Sprintf("Error streaming logs: %s\n", err)))
		}
	}()
	defer func() {
		stopLogStream()
		select {
		case <-logStreamDone:
		case <-time.After(5 * time.Second):
			u.logger.Warnf("[ExecuteJudgment] Timed out waiting for log streaming of judgment %s", j.ID())
		}
	}()

	if err := u.execLifecycles(ctx, logWriter, env, j.Lifecycle().BeforeAll()); err != nil {
		return err
	}
	if err := abortIfNoLongerJudging(u.judgmentRepo.SaveProgressIfJudging(j)); err != nil {
		return err
	}

	if err := u.runTestcases(ctx, logWriter, env, j); err != nil {
		return err
	}

	if err := u.execLifecycles(ctx, logWriter, env, j.Lifecycle().AfterAll()); err != nil {
		return err
	}
	return abortIfNoLongerJudging(u.judgmentRepo.SaveProgressIfJudging(j))
}

// provision retries environment provisioning up to provisionMaxAttempts.
// Provision cleans up after itself on failure, so each retry starts from a
// clean slate. Every failed attempt is written to the worker log so admins
// can see transient host-load trouble without digging into the database;
// cancellation and the judgment's duration budget abort the loop immediately
// (the backoff wait is context-aware).
func (u *usecase) provision(ctx context.Context, w io.Writer, spec *domainservice.EnvironmentSpec) (domainservice.ExecutionEnvironment, error) {
	for attempt := 1; ; attempt++ {
		env, err := u.provisioner.Provision(ctx, spec)
		if err == nil {
			if attempt > 1 {
				w.Write([]byte(fmt.Sprintf("Provision succeeded on attempt %d/%d\n", attempt, u.provisionMaxAttempts)))
			}
			return env, nil
		}
		if ctx.Err() != nil {
			return nil, err
		}
		if attempt >= u.provisionMaxAttempts {
			w.Write([]byte(fmt.Sprintf("Provision attempt %d/%d failed: %v — giving up\n", attempt, u.provisionMaxAttempts, err)))
			return nil, fmt.Errorf("provision failed after %d attempts: %w", attempt, err)
		}
		u.logger.Errorf("[ExecuteJudgment] Provision attempt %d/%d for judgment %s failed: %v — retrying in %s", attempt, u.provisionMaxAttempts, spec.JudgmentID, err, u.provisionRetryBackoff)
		w.Write([]byte(fmt.Sprintf("Provision attempt %d/%d failed: %v — retrying in %s\n", attempt, u.provisionMaxAttempts, err, u.provisionRetryBackoff)))
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(u.provisionRetryBackoff):
		}
	}
}

func (u *usecase) execLifecycles(ctx context.Context, w io.Writer, env domainservice.ExecutionEnvironment, lifecycles []lifecycle.Lifecycle) error {
	for i := range lifecycles {
		lc := lifecycles[i]
		w.Write([]byte(fmt.Sprintf("Running %s lifecycle\nExecution environment: %s\nCommand: %s\n", lc.Type().String(), lc.ExecutionEnvironment().String(), lc.Command().String())))

		target := domainservice.TargetTest
		if lc.ExecutionEnvironment().IsProject() {
			target = domainservice.TargetProject
		}

		report, err := env.Exec(ctx, target, lc.Command(), u.lifecycleTimeout)
		if err != nil {
			return err
		}
		if report.TimedOut {
			return fmt.Errorf("lifecycle %s timed out after %v", lc.Type().String(), u.lifecycleTimeout)
		}
		if report.OutputLost {
			return fmt.Errorf("failed to read output for lifecycle %s", lc.Type().String())
		}

		lifecycles[i].SetResult(report.Result)
		exitCode := report.Result.ExitCode()
		// Same shape as the testcase footer, so an operator scanning the
		// worker log sees the verdict of every step, not just "finished".
		w.Write([]byte(fmt.Sprintf("--- Lifecycle %s finished ---\nExit code: %d\n", lc.Type().String(), exitCode)))
		if exitCode == 0 {
			continue
		}
		// Non-zero: surface the tail of stderr right here. The field case was a
		// mysql import denied with ERROR 1044 that only ever showed up as a
		// contestant's inexplicable failing score.
		w.Write([]byte(fmt.Sprintf("Lifecycle %s stderr (tail):\n%s\n", lc.Type().String(), tailOf(report.Result.Stderr(), lifecycleStderrTailBytes))))

		if lc.Type().IsBeforeAll() || lc.Type().IsBeforeEach() {
			// before_* prepares the environment the testcases run in; a
			// failed preparation makes every later result meaningless, so the
			// judgment aborts here as ERROR instead of grading a broken setup.
			w.Write([]byte(fmt.Sprintf("Aborting judgment: lifecycle %s failed with exit code %d\n", lc.Type().String(), exitCode)))
			return &lifecycleFailedError{kind: lc.Type().String(), exitCode: exitCode}
		}
		// after_* is cleanup; a non-zero exit there is routine and tolerated.
	}
	return nil
}

// LifecycleSetupFailedReason is the contestant-facing reason attached when a
// before_* lifecycle exits non-zero. Deliberately says nothing about the
// command or its output — the worker log carries those for the operator.
const LifecycleSetupFailedReason = "評測前置步驟失敗，請聯絡現場工作人員"

// lifecycleStderrTailBytes bounds how much of a failed lifecycle's stderr is
// echoed into the worker log.
const lifecycleStderrTailBytes = 500

// lifecycleFailedError marks a before_* lifecycle that exited non-zero. Its
// message names the step and the code only: it becomes the judgment's
// errorMessage, which must not carry the command or its output either.
type lifecycleFailedError struct {
	kind     string
	exitCode int
}

func (e *lifecycleFailedError) Error() string {
	return fmt.Sprintf("lifecycle %s exited with code %d", e.kind, e.exitCode)
}

// cancelUnfinishedTestcases sweeps every testcase the abort left mid-flight —
// Pending ones never reached, and the Running one whose own before_each
// failed — to Cancelled. Without it an ERROR judgment sat over a column of
// "Pending" testcases, reading as unfinished rather than aborted. Cancelled
// is the vocabulary the context-cancel path already uses for testcases that
// stop before producing a result, so no new state reaches the UI; the
// grading item then derives FAILED from it, exactly as after a ctx cancel.
// Testcases that already finished keep their results (the operator sees what
// ran before the setup broke). This walks the whole judgment directly rather
// than relying on the run loop, so dependency-blocked testcases the scheduler
// never handed out are swept too.
func cancelUnfinishedTestcases(j *judgment.Judgment) {
	for _, gi := range j.GradingItems() {
		for _, tc := range gi.Testcases() {
			if tc.Status().IsPending() || tc.Status().IsRunning() {
				tc.SetStatus(status.CANCELLED)
			}
		}
	}
}

func tailOf(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return "..." + s[len(s)-n:]
}

func (u *usecase) runTestcases(ctx context.Context, w io.Writer, env domainservice.ExecutionEnvironment, j *judgment.Judgment) error {
	state := domainservice.NewExecutionState()
	for {
		batch := domainservice.NextBatch(j, state)
		if batch == nil {
			break
		}

		for _, tc := range batch.Testcases {
			if err := u.runTestcase(ctx, w, env, j, batch, tc, state); err != nil {
				if ctx.Err() != nil {
					// remaining testcases still get marked CANCELLED one by one
					continue
				}
				return err
			}
		}
	}
	return nil
}

func (u *usecase) runTestcase(
	ctx context.Context,
	w io.Writer,
	env domainservice.ExecutionEnvironment,
	j *judgment.Judgment,
	batch *domainservice.Batch,
	tc *testcase.Testcase,
	state *domainservice.ExecutionState,
) error {
	u.logger.Debugf("[ExecuteJudgment] Running testcase %d for judgment %s", tc.ID(), j.ID())
	tc.SetStatus(status.RUNNING)

	defer func() {
		if ctx.Err() != nil {
			tc.SetStatus(status.CANCELLED)
		}
	}()

	if err := u.execLifecycles(ctx, w, env, tc.Lifecycle().BeforeEach()); err != nil {
		return err
	}
	if err := u.saveAndPublish(j); err != nil {
		return err
	}

	w.Write([]byte(fmt.Sprintf("Running testcase %d\nCommand: %s\n", tc.ID(), tc.Command().String())))

	report, err := env.Exec(ctx, domainservice.TargetTest, tc.Command(), u.testcaseTimeout)
	if err != nil {
		if ctx.Err() != nil {
			// deferred cleanup marks this testcase CANCELLED; persist that
			tc.SetStatus(status.CANCELLED)
			u.saveAndNotify(j)
		}
		return err
	}

	if report.TimedOut {
		w.Write([]byte(fmt.Sprintf("Testcase %d timed out\n", tc.ID())))
		tc.SetStatus(status.TIMEOUT)
	} else if report.OutputLost {
		u.logger.Errorf("[ExecuteJudgment] Marking testcase %d as ERROR due to output read failure in judgment %s", tc.ID(), j.ID())
		tc.SetStatus(status.ERROR)
	}

	tc.SetResult(report.Result)

	if !report.TimedOut && !report.OutputLost {
		gradingItem := batch.GradingItemFor(tc.ID())
		tcStatus, parsedScore, publicStdout := gradingItem.ParseResult(tc.Result(), tc.Score())
		if publicStdout != tc.Result().Stdout() {
			// The parser consumed its verdict line — persist only what
			// remains, so the grader's control output never reaches the
			// contestant-visible sample-testcase stdout.
			tc.SetResult(execresult.NewExecResult(&execresult.ExecResultConfig{
				ExitCode: tc.Result().ExitCode(),
				ExecTime: tc.Result().ExecTime(),
				Stdout:   publicStdout,
				Stderr:   tc.Result().Stderr(),
			}))
		}
		tc.SetStatus(tcStatus)
		tc.SetParsedScore(parsedScore)
		if tcStatus.IsPassed() {
			state.MarkPassed(tc.ID())
		}
		u.logger.Debugf("[ExecuteJudgment] Testcase %d status %s for judgment %s", tc.ID(), tcStatus.String(), j.ID())
	}

	if err := u.saveAndPublish(j); err != nil {
		return err
	}
	w.Write([]byte(fmt.Sprintf("Testcase %d finished\nExit code: %d\nExecution time: %dms\n--- Testcase %d finished ---\n",
		tc.ID(), report.Result.ExitCode(), report.Result.ExecTime(), tc.ID())))

	if err := u.execLifecycles(ctx, w, env, tc.Lifecycle().AfterEach()); err != nil {
		return err
	}

	return abortIfNoLongerJudging(u.judgmentRepo.SaveProgressIfJudging(j))
}
