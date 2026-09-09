package executejudgment_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"sync"
	"testing"
	"time"

	"judge-backend/internal/app/judge_domain/events/judged"
	"judge-backend/internal/app/judge_domain/events/judgeupdated"
	executejudgment "judge-backend/internal/app/judge_domain/judge/application/usecase/execute_judgment"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/command"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/execresult"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/gradingitem"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/gradingitem/scoringstrategy"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/judgetype"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/lifecycle"
	judgmentproject "judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/project"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/testcase"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/testcase/resultparser"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/submission"
	domainservice "judge-backend/internal/app/judge_domain/judge/domain/service"
	judgemocks "judge-backend/internal/app/judge_domain/judge/mocks"
	sharedeventbus "judge-backend/internal/app/shared/infrastructure/eventbus"
	sharedlogger "judge-backend/internal/app/shared/infrastructure/logger"
	sharedtypes "judge-backend/internal/app/shared/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// syncWriter guards the log buffer: the usecase writes progress messages from
// the main flow while StreamLogs may write from its own goroutine.
type syncWriter struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (w *syncWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.buf.Write(p)
}

// String makes fmt (and testify's mock-argument formatting, which runs on the
// StreamLogs goroutine) read the buffer under the mutex instead of reflecting
// into it unsynchronized.
func (w *syncWriter) String() string {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.buf.String()
}

func newRunnableJudgment(jt judgetype.JudgeType) *judgment.Judgment {
	tc := testcase.NewTestcase(&testcase.TestcaseConfig{ID: 1, Command: "echo hello", Score: 100})
	tc.SetLifecycle(lifecycle.NewTestcaseLifecycle(nil))
	gi := gradingitem.NewGradingItem(&gradingitem.GradingItemConfig{
		ID:              1,
		Score:           100,
		ScoringStrategy: scoringstrategy.IncreaseScoreStrategy,
		Testcases:       []*testcase.Testcase{tc},
	})
	j := judgment.New(&judgment.JudgmentConfig{
		ContestProjectID: 1,
		SubmissionID:     "sub-1",
		Project:          judgmentproject.NewProject(&judgmentproject.ProjectConfig{ID: 1}),
		GradingItems:     []*gradingitem.GradingItem{gi},
		Lifecycle:        lifecycle.NewJudgmentLifecycle(nil),
		JudgeType:        jt,
	})
	j.SetID("judgment-1")
	return j
}

func passingReport() *domainservice.ExecReport {
	return &domainservice.ExecReport{
		Result: execresult.NewExecResult(&execresult.ExecResultConfig{ExitCode: 0, ExecTime: 1}),
	}
}

type deps struct {
	provisioner       *judgemocks.ExecutionEnvironmentProvisioner
	environment       *judgemocks.ExecutionEnvironment
	submissionService *judgemocks.SubmissionService
	submissionRepo    *judgemocks.SubmissionRepository
	userRepo          *judgemocks.UserRepository
	judgmentRepo      *judgemocks.JudgmentRepository
}

func newDeps(t *testing.T) *deps {
	return &deps{
		provisioner:       judgemocks.NewExecutionEnvironmentProvisioner(t),
		environment:       judgemocks.NewExecutionEnvironment(t),
		submissionService: judgemocks.NewSubmissionService(t),
		submissionRepo:    judgemocks.NewSubmissionRepository(t),
		userRepo:          judgemocks.NewUserRepository(t),
		judgmentRepo:      judgemocks.NewJudgmentRepository(t),
	}
}

func buildUsecase(d *deps, eb sharedtypes.EventBus, logger sharedtypes.Logger) executejudgment.Usecase {
	// provision retry backoff 0: unit tests exercise the retry LOGIC, not the pause
	return executejudgment.NewUsecase(d.provisioner, d.submissionService, d.submissionRepo, d.userRepo, d.judgmentRepo, eb, 60*time.Second, 120*time.Second, 3, 0, logger)
}

func newUsecase(t *testing.T) (executejudgment.Usecase, *deps) {
	d := newDeps(t)
	logger := sharedlogger.NewMockLogger()
	eb := sharedeventbus.NewEventBus(logger)
	return buildUsecase(d, eb, logger), d
}

func expectStreamLogs(d *deps) {
	d.environment.On("StreamLogs", mock.Anything, mock.Anything).Return(func(ctx context.Context, _ io.Writer) error {
		<-ctx.Done()
		return ctx.Err()
	}).Maybe()
}

// expectSubmitterLookup satisfies the submission → username resolution the
// usecase performs before provisioning ("sub-1" is owned by user 7, "web07").
func expectSubmitterLookup(d *deps) {
	sub, _ := submission.New(&submission.SubmissionConfig{UserID: 7})
	sub.SetID("sub-1")
	d.submissionRepo.On("FindByID", "sub-1").Return(sub, nil).Once()
	d.userRepo.On("FindUsernameByID", int32(7)).Return("web07", nil).Once()
}

func TestExecute_HappyPath_FinishesJudged(t *testing.T) {
	u, d := newUsecase(t)
	j := newRunnableJudgment(judgetype.JUDGE)

	d.submissionService.On("LoadMetadata", "sub-1").Return(map[string]string{}, nil).Once()
	expectSubmitterLookup(d)
	d.provisioner.On("Provision", mock.Anything, mock.MatchedBy(func(spec *domainservice.EnvironmentSpec) bool {
		return spec.JudgmentID == "judgment-1" && spec.SubmissionID == "sub-1" && spec.WorkerID == "worker-1" &&
			spec.SubmitterUsername == "web07"
	})).Return(d.environment, nil).Once()
	expectStreamLogs(d)
	d.environment.On("Exec", mock.Anything, domainservice.TargetTest, mock.Anything, mock.Anything).Return(passingReport(), nil).Once()
	d.environment.On("Teardown").Once()
	d.judgmentRepo.On("TryStart", j.ID(), mock.AnythingOfType("time.Time")).Return(true, nil)
	d.judgmentRepo.On("SaveProgressIfJudging", j).Return(true, nil)
	d.judgmentRepo.On("SaveTerminalIfJudging", j).Return(true, nil)

	err := u.Execute(context.Background(), &syncWriter{}, j, "worker-1")

	assert.NoError(t, err)
	assert.True(t, j.Status().IsJudged(), "expected JUDGED, got %s", j.Status().String())
	tc := j.GradingItems()[0].Testcases()[0]
	assert.True(t, tc.Status().IsPassed(), "expected PASSED, got %s", tc.Status().String())
	assert.Equal(t, 100.0, j.CalculateScore())
}

func TestExecute_RejudgeType_FinishesRejudged(t *testing.T) {
	u, d := newUsecase(t)
	j := newRunnableJudgment(judgetype.REJUDGE)

	d.submissionService.On("LoadMetadata", "sub-1").Return(map[string]string{}, nil).Once()
	expectSubmitterLookup(d)
	d.provisioner.On("Provision", mock.Anything, mock.Anything).Return(d.environment, nil).Once()
	expectStreamLogs(d)
	d.environment.On("Exec", mock.Anything, domainservice.TargetTest, mock.Anything, mock.Anything).Return(passingReport(), nil).Once()
	d.environment.On("Teardown").Once()
	d.judgmentRepo.On("TryStart", j.ID(), mock.AnythingOfType("time.Time")).Return(true, nil)
	d.judgmentRepo.On("SaveProgressIfJudging", j).Return(true, nil)
	d.judgmentRepo.On("SaveTerminalIfJudging", j).Return(true, nil)

	err := u.Execute(context.Background(), &syncWriter{}, j, "worker-1")

	assert.NoError(t, err)
	assert.True(t, j.Status().IsRejudged(), "expected REJUDGED, got %s", j.Status().String())
}

func TestExecute_ProvisionFailure_RetriesThenFinishesError(t *testing.T) {
	u, d := newUsecase(t)
	j := newRunnableJudgment(judgetype.JUDGE)

	d.submissionService.On("LoadMetadata", "sub-1").Return(map[string]string{}, nil).Once()
	expectSubmitterLookup(d)
	// every attempt fails — the loop must try exactly provisionMaxAttempts times
	d.provisioner.On("Provision", mock.Anything, mock.Anything).Return(nil, errors.New("no runtime capacity")).Times(3)
	d.judgmentRepo.On("TryStart", j.ID(), mock.AnythingOfType("time.Time")).Return(true, nil)
	d.judgmentRepo.On("SaveTerminalIfJudging", j).Return(true, nil)

	w := &syncWriter{}
	err := u.Execute(context.Background(), w, j, "worker-1")

	assert.Error(t, err)
	assert.True(t, j.Status().IsError(), "expected ERROR, got %s", j.Status().String())
	assert.NotNil(t, j.JudgingFinishedAt())
	if assert.NotNil(t, j.ErrorMessage()) {
		assert.Contains(t, *j.ErrorMessage(), "after 3 attempts", "the persisted error must say retries were exhausted")
	}
	assert.Contains(t, w.String(), "Provision attempt 1/3 failed", "each failed attempt must reach the worker log")
	assert.Contains(t, w.String(), "Provision attempt 3/3 failed", "the giving-up attempt must reach the worker log")
}

func TestExecute_ProvisionRetry_SucceedsOnSecondAttempt(t *testing.T) {
	u, d := newUsecase(t)
	j := newRunnableJudgment(judgetype.JUDGE)

	d.submissionService.On("LoadMetadata", "sub-1").Return(map[string]string{}, nil).Once()
	expectSubmitterLookup(d)
	d.provisioner.On("Provision", mock.Anything, mock.Anything).Return(nil, errors.New("project container became unhealthy")).Once()
	d.provisioner.On("Provision", mock.Anything, mock.Anything).Return(d.environment, nil).Once()
	expectStreamLogs(d)
	d.environment.On("Exec", mock.Anything, domainservice.TargetTest, mock.Anything, mock.Anything).Return(passingReport(), nil).Once()
	d.environment.On("Teardown").Once()
	d.judgmentRepo.On("TryStart", j.ID(), mock.AnythingOfType("time.Time")).Return(true, nil)
	d.judgmentRepo.On("SaveProgressIfJudging", j).Return(true, nil)
	d.judgmentRepo.On("SaveTerminalIfJudging", j).Return(true, nil)

	w := &syncWriter{}
	err := u.Execute(context.Background(), w, j, "worker-1")

	assert.NoError(t, err)
	assert.True(t, j.Status().IsJudged(), "expected JUDGED after a successful retry, got %s", j.Status().String())
	assert.Contains(t, w.String(), "Provision attempt 1/3 failed")
	assert.Contains(t, w.String(), "Provision succeeded on attempt 2/3")
}

func TestExecute_ProvisionRetry_CancelledMidLoop_FinishesCancelled(t *testing.T) {
	u, d := newUsecase(t)
	j := newRunnableJudgment(judgetype.JUDGE)

	ctx, cancel := context.WithCancel(context.Background())

	d.submissionService.On("LoadMetadata", "sub-1").Return(map[string]string{}, nil).Once()
	expectSubmitterLookup(d)
	// cancellation lands during the first failed attempt — no further retries
	d.provisioner.On("Provision", mock.Anything, mock.Anything).
		Run(func(mock.Arguments) { cancel() }).
		Return(nil, errors.New("project container became unhealthy")).Once()
	d.judgmentRepo.On("TryStart", j.ID(), mock.AnythingOfType("time.Time")).Return(true, nil)
	d.judgmentRepo.On("SaveTerminalIfJudging", j).Return(true, nil)

	_ = u.Execute(ctx, &syncWriter{}, j, "worker-1")

	assert.True(t, j.Status().IsCancelled(), "expected CANCELLED, got %s", j.Status().String())
}

func TestExecute_MetadataFailure_FinishesError(t *testing.T) {
	u, d := newUsecase(t)
	j := newRunnableJudgment(judgetype.JUDGE)

	d.submissionService.On("LoadMetadata", "sub-1").Return(nil, errors.New("metadata.json not found")).Once()
	d.judgmentRepo.On("TryStart", j.ID(), mock.AnythingOfType("time.Time")).Return(true, nil)
	d.judgmentRepo.On("SaveTerminalIfJudging", j).Return(true, nil)

	err := u.Execute(context.Background(), &syncWriter{}, j, "worker-1")

	assert.Error(t, err)
	assert.True(t, j.Status().IsError(), "expected ERROR, got %s", j.Status().String())
}

// TestExecute_StdoutVerdictLine_StrippedFromPersistedResult pins the
// contestant-facing contract: sample-testcase stdout is visible to
// contestants, so the stdout parser's "pass/fail {score}" control line must
// be consumed at parse time and never reach the persisted result — while the
// grading script's diagnostics after it (and stderr) survive untouched.
func TestExecute_StdoutVerdictLine_StrippedFromPersistedResult(t *testing.T) {
	u, d := newUsecase(t)

	tc := testcase.NewTestcase(&testcase.TestcaseConfig{ID: 1, Command: "echo verdict", Score: 1})
	tc.SetLifecycle(lifecycle.NewTestcaseLifecycle(nil))
	gi := gradingitem.NewGradingItem(&gradingitem.GradingItemConfig{
		ID:              1,
		Score:           1,
		ScoringStrategy: scoringstrategy.IncreaseScoreStrategy,
		ResultParser:    resultparser.StdoutParser,
		Testcases:       []*testcase.Testcase{tc},
	})
	j := judgment.New(&judgment.JudgmentConfig{
		ContestProjectID: 1,
		SubmissionID:     "sub-1",
		Project:          judgmentproject.NewProject(&judgmentproject.ProjectConfig{ID: 1}),
		GradingItems:     []*gradingitem.GradingItem{gi},
		Lifecycle:        lifecycle.NewJudgmentLifecycle(nil),
		JudgeType:        judgetype.JUDGE,
	})
	j.SetID("judgment-1")

	d.submissionService.On("LoadMetadata", "sub-1").Return(map[string]string{}, nil).Once()
	expectSubmitterLookup(d)
	d.provisioner.On("Provision", mock.Anything, mock.Anything).Return(d.environment, nil).Once()
	expectStreamLogs(d)
	d.environment.On("Exec", mock.Anything, domainservice.TargetTest, mock.Anything, mock.Anything).Return(&domainservice.ExecReport{
		Result: execresult.NewExecResult(&execresult.ExecResultConfig{
			ExitCode: 0,
			ExecTime: 1,
			Stdout:   "pass 0.75\nselector <div> found",
			Stderr:   "diagnostic stderr",
		}),
	}, nil).Once()
	d.environment.On("Teardown").Once()
	d.judgmentRepo.On("TryStart", j.ID(), mock.AnythingOfType("time.Time")).Return(true, nil)
	d.judgmentRepo.On("SaveProgressIfJudging", j).Return(true, nil)
	d.judgmentRepo.On("SaveTerminalIfJudging", j).Return(true, nil)

	err := u.Execute(context.Background(), &syncWriter{}, j, "worker-1")

	assert.NoError(t, err)
	got := j.GradingItems()[0].Testcases()[0]
	assert.True(t, got.Status().IsPassed(), "expected PASSED, got %s", got.Status().String())
	assert.Equal(t, "selector <div> found", got.Result().Stdout(), "verdict line must be stripped from the persisted stdout")
	assert.Equal(t, "diagnostic stderr", got.Result().Stderr(), "stderr must be untouched")
	assert.Equal(t, 0.75, j.CalculateScore(), "the stdout-provided score must still be applied")
}

func TestExecute_SubmitterLookupFailure_FinishesError(t *testing.T) {
	u, d := newUsecase(t)
	j := newRunnableJudgment(judgetype.JUDGE)

	d.submissionService.On("LoadMetadata", "sub-1").Return(map[string]string{}, nil).Once()
	d.submissionRepo.On("FindByID", "sub-1").Return(nil, errors.New("mongo unavailable")).Once()
	d.judgmentRepo.On("TryStart", j.ID(), mock.AnythingOfType("time.Time")).Return(true, nil)
	d.judgmentRepo.On("SaveTerminalIfJudging", j).Return(true, nil)

	err := u.Execute(context.Background(), &syncWriter{}, j, "worker-1")

	assert.Error(t, err)
	assert.True(t, j.Status().IsError(), "expected ERROR, got %s", j.Status().String())
}

func TestExecute_UsernameLookupFailure_FinishesError(t *testing.T) {
	u, d := newUsecase(t)
	j := newRunnableJudgment(judgetype.JUDGE)

	sub, _ := submission.New(&submission.SubmissionConfig{UserID: 7})
	sub.SetID("sub-1")
	d.submissionService.On("LoadMetadata", "sub-1").Return(map[string]string{}, nil).Once()
	d.submissionRepo.On("FindByID", "sub-1").Return(sub, nil).Once()
	d.userRepo.On("FindUsernameByID", int32(7)).Return("", errors.New("user table unavailable")).Once()
	d.judgmentRepo.On("TryStart", j.ID(), mock.AnythingOfType("time.Time")).Return(true, nil)
	d.judgmentRepo.On("SaveTerminalIfJudging", j).Return(true, nil)

	err := u.Execute(context.Background(), &syncWriter{}, j, "worker-1")

	assert.Error(t, err)
	assert.True(t, j.Status().IsError(), "expected ERROR, got %s", j.Status().String())
}

func TestExecute_CancelledContext_FinishesCancelled(t *testing.T) {
	u, d := newUsecase(t)
	j := newRunnableJudgment(judgetype.JUDGE)

	ctx, cancel := context.WithCancel(context.Background())

	d.submissionService.On("LoadMetadata", "sub-1").Return(map[string]string{}, nil).Once()
	expectSubmitterLookup(d)
	d.provisioner.On("Provision", mock.Anything, mock.Anything).Return(d.environment, nil).Once()
	expectStreamLogs(d)
	d.environment.On("Exec", mock.Anything, domainservice.TargetTest, mock.Anything, mock.Anything).
		Return(func(ctx context.Context, _ domainservice.ExecTarget, _ command.Command, _ time.Duration) (*domainservice.ExecReport, error) {
			cancel()
			<-ctx.Done()
			return nil, ctx.Err()
		}).Once()
	d.environment.On("Teardown").Once()
	d.judgmentRepo.On("TryStart", j.ID(), mock.AnythingOfType("time.Time")).Return(true, nil)
	d.judgmentRepo.On("SaveProgressIfJudging", j).Return(true, nil)
	d.judgmentRepo.On("SaveTerminalIfJudging", j).Return(true, nil)

	// the run drains remaining testcases as CANCELLED instead of erroring out,
	// so Execute's error value is not asserted here — the terminal status is
	_ = u.Execute(ctx, &syncWriter{}, j, "worker-1")

	assert.True(t, j.Status().IsCancelled(), "expected CANCELLED, got %s", j.Status().String())
	tc := j.GradingItems()[0].Testcases()[0]
	assert.True(t, tc.Status().IsCancelled(), "expected testcase CANCELLED, got %s", tc.Status().String())
}

func TestExecute_DeadlineExceeded_FinishesErrorNotCancelled(t *testing.T) {
	u, d := newUsecase(t)
	j := newRunnableJudgment(judgetype.JUDGE)

	// emulate the worker's emergency budget expiring mid-run
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	d.submissionService.On("LoadMetadata", "sub-1").Return(map[string]string{}, nil).Once()
	expectSubmitterLookup(d)
	d.provisioner.On("Provision", mock.Anything, mock.Anything).Return(d.environment, nil).Once()
	expectStreamLogs(d)
	d.environment.On("Exec", mock.Anything, domainservice.TargetTest, mock.Anything, mock.Anything).
		Return(func(ctx context.Context, _ domainservice.ExecTarget, _ command.Command, _ time.Duration) (*domainservice.ExecReport, error) {
			<-ctx.Done()
			return nil, ctx.Err()
		}).Once()
	d.environment.On("Teardown").Once()
	d.judgmentRepo.On("TryStart", j.ID(), mock.AnythingOfType("time.Time")).Return(true, nil)
	d.judgmentRepo.On("SaveProgressIfJudging", j).Return(true, nil)
	d.judgmentRepo.On("SaveTerminalIfJudging", j).Return(true, nil)

	_ = u.Execute(ctx, &syncWriter{}, j, "worker-1")

	assert.True(t, j.Status().IsError(), "expected ERROR on deadline, got %s", j.Status().String())
	assert.NotNil(t, j.ErrorMessage(), "expected error message to be recorded")
}

func TestExecute_TimedOutTestcase_MarksTimeoutAndFinishes(t *testing.T) {
	u, d := newUsecase(t)
	j := newRunnableJudgment(judgetype.JUDGE)

	d.submissionService.On("LoadMetadata", "sub-1").Return(map[string]string{}, nil).Once()
	expectSubmitterLookup(d)
	d.provisioner.On("Provision", mock.Anything, mock.Anything).Return(d.environment, nil).Once()
	expectStreamLogs(d)
	d.environment.On("Exec", mock.Anything, domainservice.TargetTest, mock.Anything, mock.Anything).Return(&domainservice.ExecReport{
		Result:   execresult.NewExecResult(&execresult.ExecResultConfig{ExitCode: -1, ExecTime: 60000}),
		TimedOut: true,
	}, nil).Once()
	d.environment.On("Teardown").Once()
	d.judgmentRepo.On("TryStart", j.ID(), mock.AnythingOfType("time.Time")).Return(true, nil)
	d.judgmentRepo.On("SaveProgressIfJudging", j).Return(true, nil)
	d.judgmentRepo.On("SaveTerminalIfJudging", j).Return(true, nil)

	err := u.Execute(context.Background(), &syncWriter{}, j, "worker-1")

	assert.NoError(t, err)
	assert.True(t, j.Status().IsJudged())
	tc := j.GradingItems()[0].Testcases()[0]
	assert.True(t, tc.Status().IsTimeout(), "expected TIMEOUT, got %s", tc.Status().String())
	assert.Equal(t, 0.0, j.CalculateScore())
}

func TestExecute_NonQueuedJudgment_IsRejected(t *testing.T) {
	u, _ := newUsecase(t)
	j := newRunnableJudgment(judgetype.JUDGE)
	assert.NoError(t, j.Cancel("already cancelled"))

	err := u.Execute(context.Background(), &syncWriter{}, j, "worker-1")

	assert.Error(t, err)
	assert.True(t, j.Status().IsCancelled(), "status must stay CANCELLED")
}

// TestExecute_TryStartLosesRace_AbortsWithoutRunning is a regression test
// for the most severe scenario in finding 1: a judgment can be cancelled by
// the API while it is still sitting in the in-memory queue, never yet popped
// by any worker — judge_cancelled's handler found no worker executing it, so
// there was nothing to interrupt, and TryCancel already moved the DB
// document straight from Queued to Cancelled. When a worker then pops that
// same (now stale) queue entry, j.Start() only validates against the
// worker's own in-memory copy — which still reads Queued — so it succeeds.
// Only TryStart's conditional DB write can catch that the document is no
// longer actually Queued. This test asserts that when TryStart reports
// applied=false, Execute aborts immediately: it must not run the judgment,
// must not publish anything, and must not attempt any further repository
// call (registering no other mock expectations means any such call would
// panic).
func TestExecute_TryStartLosesRace_AbortsWithoutRunning(t *testing.T) {
	u, d := newUsecase(t)
	j := newRunnableJudgment(judgetype.JUDGE)

	d.judgmentRepo.On("TryStart", j.ID(), mock.AnythingOfType("time.Time")).Return(false, nil).Once()

	err := u.Execute(context.Background(), &syncWriter{}, j, "worker-1")

	assert.Error(t, err)
}

// TestExecute_TerminalTransitionFails_FailsClosedWithoutSaveOrPublish covers
// Execute's tail: each branch there assumes j is still Judging and moves it
// to the branch's terminal status via Fail/Cancel/Finish. If that domain
// transition is rejected, j is not in the state the branch assumed, and the
// old behavior — log the rejection, then save-and-publish anyway — announced
// an outcome nobody computed. Execute must instead return the transition
// error without calling SaveTerminalIfJudging (no mock expectation is
// registered for it, so any call panics) and without publishing judged.
//
// The mid-run mutation is simulated from the mocked Exec callback, which
// moves j to Cancelled before failing — so the runErr branch's j.Fail()
// is rejected (Cancelled is terminal).
func TestExecute_TerminalTransitionFails_FailsClosedWithoutSaveOrPublish(t *testing.T) {
	d := newDeps(t)
	logger := sharedlogger.NewMockLogger()
	eb := sharedeventbus.NewEventBus(logger)
	u := buildUsecase(d, eb, logger)

	judgedPublished := false
	_, subErr := eb.Subscribe(func(e *judged.JudgedEvent) {
		judgedPublished = true
	})
	assert.NoError(t, subErr)

	j := newRunnableJudgment(judgetype.JUDGE)

	d.submissionService.On("LoadMetadata", "sub-1").Return(map[string]string{}, nil).Once()
	expectSubmitterLookup(d)
	d.provisioner.On("Provision", mock.Anything, mock.Anything).Return(d.environment, nil).Once()
	expectStreamLogs(d)
	d.environment.On("Exec", mock.Anything, domainservice.TargetTest, mock.Anything, mock.Anything).
		Return(func(ctx context.Context, _ domainservice.ExecTarget, _ command.Command, _ time.Duration) (*domainservice.ExecReport, error) {
			assert.NoError(t, j.Cancel("mutated out of band mid-run"))
			return nil, errors.New("exec failed after the entity left Judging")
		}).Once()
	d.environment.On("Teardown").Once()
	d.judgmentRepo.On("TryStart", j.ID(), mock.AnythingOfType("time.Time")).Return(true, nil)
	d.judgmentRepo.On("SaveProgressIfJudging", j).Return(true, nil)
	// Deliberately NO SaveTerminalIfJudging / FindByID expectations.

	err := u.Execute(context.Background(), &syncWriter{}, j, "worker-1")

	assert.Error(t, err, "a rejected terminal transition must surface as an error")
	assert.False(t, judgedPublished, "must not publish judged for an outcome that was never computed")
}

// TestExecute_TerminalSaveLosesRace_PublishesActualDBState is a regression
// test: the terminal save used to be a plain, unconditional Save — whichever
// writer's Save landed last simply won. An independent cancel-request
// handler can save CANCELLED to the same judgment ID while this worker is
// still executing; if the worker then finished normally and saved JUDGED
// unconditionally, that would silently discard the already-accepted
// cancellation. SaveTerminalIfJudging closes that by only applying when the
// judgment is still Judging; this test drives the "lost the race" path
// directly (SaveTerminalIfJudging returns applied=false) and asserts the
// published judge_updated event reflects the re-fetched actual DB state
// (Cancelled), not the worker's own stale, overridden computation (Judged).
func TestExecute_TerminalSaveLosesRace_PublishesActualDBState(t *testing.T) {
	d := newDeps(t)
	logger := sharedlogger.NewMockLogger()
	eb := sharedeventbus.NewEventBus(logger)
	u := buildUsecase(d, eb, logger)

	var lastPublishedStatus string
	_, subErr := eb.Subscribe(func(e *judgeupdated.JudgeUpdatedEvent) {
		lastPublishedStatus = e.Data().Status
	})
	assert.NoError(t, subErr)

	j := newRunnableJudgment(judgetype.JUDGE)

	d.submissionService.On("LoadMetadata", "sub-1").Return(map[string]string{}, nil).Once()
	expectSubmitterLookup(d)
	d.provisioner.On("Provision", mock.Anything, mock.Anything).Return(d.environment, nil).Once()
	expectStreamLogs(d)
	d.environment.On("Exec", mock.Anything, domainservice.TargetTest, mock.Anything, mock.Anything).Return(passingReport(), nil).Once()
	d.environment.On("Teardown").Once()
	d.judgmentRepo.On("TryStart", j.ID(), mock.AnythingOfType("time.Time")).Return(true, nil)
	d.judgmentRepo.On("SaveProgressIfJudging", j).Return(true, nil)
	d.judgmentRepo.On("SaveTerminalIfJudging", j).Return(false, nil)

	cancelledInDB := newRunnableJudgment(judgetype.JUDGE)
	assert.NoError(t, cancelledInDB.Cancel("user requested cancellation"))
	d.judgmentRepo.On("FindByID", "judgment-1").Return(cancelledInDB, nil).Once()

	err := u.Execute(context.Background(), &syncWriter{}, j, "worker-1")

	assert.NoError(t, err)
	// The caller's own j reference is unaffected by Execute's internal
	// reassignment after losing the race — it still shows the worker's own
	// (now-overridden) computed outcome.
	assert.True(t, j.Status().IsJudged())
	assert.Equal(t, "Cancelled", lastPublishedStatus, "expected the published update to reflect the actual (re-fetched) DB state, not the worker's stale computed result")
}

// TestExecute_TerminalSaveLosesRaceAndRefetchFails_NoPublish is a regression
// test for finding 4: when the terminal save loses the race AND the
// follow-up re-fetch of the actual DB state also fails, there is no
// confirmed persisted state left to publish at all. The old code published
// the worker's own stale computed outcome (a judged event) regardless; this
// test asserts Execute instead returns an error and never publishes a judged
// event (mid-run judge_updated progress publishes, which reflect state that
// really was persisted via SaveProgressIfJudging, are unrelated and still
// expected).
func TestExecute_TerminalSaveLosesRaceAndRefetchFails_NoPublish(t *testing.T) {
	d := newDeps(t)
	logger := sharedlogger.NewMockLogger()
	eb := sharedeventbus.NewEventBus(logger)
	u := buildUsecase(d, eb, logger)

	judgedPublished := false
	_, subErr := eb.Subscribe(func(e *judged.JudgedEvent) {
		judgedPublished = true
	})
	assert.NoError(t, subErr)

	j := newRunnableJudgment(judgetype.JUDGE)

	d.submissionService.On("LoadMetadata", "sub-1").Return(map[string]string{}, nil).Once()
	expectSubmitterLookup(d)
	d.provisioner.On("Provision", mock.Anything, mock.Anything).Return(d.environment, nil).Once()
	expectStreamLogs(d)
	d.environment.On("Exec", mock.Anything, domainservice.TargetTest, mock.Anything, mock.Anything).Return(passingReport(), nil).Once()
	d.environment.On("Teardown").Once()
	d.judgmentRepo.On("TryStart", j.ID(), mock.AnythingOfType("time.Time")).Return(true, nil)
	d.judgmentRepo.On("SaveProgressIfJudging", j).Return(true, nil)
	d.judgmentRepo.On("SaveTerminalIfJudging", j).Return(false, nil)
	d.judgmentRepo.On("FindByID", "judgment-1").Return(nil, errors.New("mongo unavailable")).Once()

	err := u.Execute(context.Background(), &syncWriter{}, j, "worker-1")

	assert.Error(t, err, "expected an error when the terminal state could not be confirmed persisted")
	assert.False(t, judgedPublished, "must not publish a judged event when the terminal state was never confirmed persisted")
}

// --- lifecycle exit codes -------------------------------------------------
//
// A before_* lifecycle that exits non-zero used to be ignored: the exit code
// was stored, the log said "finished", and judging carried on against an
// environment that was never set up (the field case: a mysql import denied
// with ERROR 1044, followed by an inexplicable failing score). before_* is now
// fail-fast with a neutral contestant-facing reason; after_* stays tolerant
// (cleanup commands are routinely non-zero) but the exit code is logged.

func lifecycleOf(kind, cmd string) *lifecycle.Lifecycle {
	return lifecycle.NewLifecycle(&lifecycle.LifecycleConfig{ID: 9, ExecutionEnvironment: "test", Type: kind, Command: cmd})
}

func execOf(cmd string) interface{} {
	return mock.MatchedBy(func(c command.Command) bool { return c.String() == cmd })
}

func failingReport(code int, stderr string) *domainservice.ExecReport {
	return &domainservice.ExecReport{
		Result: execresult.NewExecResult(&execresult.ExecResultConfig{ExitCode: code, ExecTime: 1, Stderr: stderr}),
	}
}

func newJudgmentWithLifecycles(judgmentLCs []*lifecycle.Lifecycle, testcaseLCs []*lifecycle.Lifecycle) *judgment.Judgment {
	tc := testcase.NewTestcase(&testcase.TestcaseConfig{ID: 1, Command: "echo hello", Score: 100})
	tc.SetLifecycle(lifecycle.NewTestcaseLifecycle(testcaseLCs))
	gi := gradingitem.NewGradingItem(&gradingitem.GradingItemConfig{
		ID: 1, Score: 100, ScoringStrategy: scoringstrategy.IncreaseScoreStrategy, Testcases: []*testcase.Testcase{tc},
	})
	j := judgment.New(&judgment.JudgmentConfig{
		ContestProjectID: 1, SubmissionID: "sub-1",
		Project:      judgmentproject.NewProject(&judgmentproject.ProjectConfig{ID: 1}),
		GradingItems: []*gradingitem.GradingItem{gi},
		Lifecycle:    lifecycle.NewJudgmentLifecycle(judgmentLCs),
		JudgeType:    judgetype.JUDGE,
	})
	j.SetID("judgment-1")
	return j
}

func expectRunSetup(d *deps) {
	d.submissionService.On("LoadMetadata", "sub-1").Return(map[string]string{}, nil).Once()
	expectSubmitterLookup(d)
	d.provisioner.On("Provision", mock.Anything, mock.Anything).Return(d.environment, nil).Once()
	expectStreamLogs(d)
	d.environment.On("Teardown").Once()
	d.judgmentRepo.On("TryStart", "judgment-1", mock.AnythingOfType("time.Time")).Return(true, nil)
	d.judgmentRepo.On("SaveTerminalIfJudging", mock.Anything).Return(true, nil)
}

func TestExecute_BeforeAllNonZero_FailsJudgmentWithNeutralReason(t *testing.T) {
	u, d := newUsecase(t)
	const setup = "mysql -h db -u web07 -p1234 < /files/init_sql"
	j := newJudgmentWithLifecycles([]*lifecycle.Lifecycle{lifecycleOf("before_all", setup)}, nil)

	expectRunSetup(d)
	d.environment.On("Exec", mock.Anything, domainservice.TargetTest, execOf(setup), mock.Anything).
		Return(failingReport(1, "ERROR 1044 (42000): Access denied for user 'web07'@'%' to database 'web07_db'"), nil).Once()
	// No Exec expectation for the testcase itself: reaching it would trip the
	// strict mock, which is exactly the abort this test is about.

	log := &syncWriter{}
	err := u.Execute(context.Background(), log, j, "worker-1")

	assert.Error(t, err)
	assert.True(t, j.Status().IsError(), "expected ERROR, got %s", j.Status().String())
	if assert.NotNil(t, j.Reason()) {
		assert.Equal(t, executejudgment.LifecycleSetupFailedReason, *j.Reason())
	}
	if assert.NotNil(t, j.ErrorMessage()) {
		assert.Contains(t, *j.ErrorMessage(), "before_all")
		assert.NotContains(t, *j.ErrorMessage(), "mysql", "the error message must not leak the command")
		assert.NotContains(t, *j.ErrorMessage(), "1044", "the error message must not leak the output")
	}
	// The worker log is where the operator sees WHY.
	assert.Contains(t, log.String(), "Exit code: 1")
	assert.Contains(t, log.String(), "ERROR 1044")
	// Nothing ran, so nothing may look like it is still going to: every
	// testcase is swept to Cancelled (the ctx-cancel vocabulary), and the
	// grading item derives FAILED from that — the same shape a ctx cancel
	// leaves behind today, so the UI needs no new state.
	tc := j.GradingItems()[0].Testcases()[0]
	assert.True(t, tc.Status().IsCancelled(), "expected the unrun testcase to be Cancelled, got %s", tc.Status().String())
	assert.True(t, j.GradingItems()[0].Status().IsFailed(), "grading item over cancelled testcases derives FAILED, got %s", j.GradingItems()[0].Status().String())
}

func TestExecute_AfterEachNonZero_StillFinishesAndLogsExitCode(t *testing.T) {
	u, d := newUsecase(t)
	const cleanup = "rm -rf /tmp/scratch"
	j := newJudgmentWithLifecycles(nil, []*lifecycle.Lifecycle{lifecycleOf("after_each", cleanup)})

	expectRunSetup(d)
	d.judgmentRepo.On("SaveProgressIfJudging", mock.Anything).Return(true, nil)
	d.environment.On("Exec", mock.Anything, domainservice.TargetTest, execOf("echo hello"), mock.Anything).Return(passingReport(), nil).Once()
	d.environment.On("Exec", mock.Anything, domainservice.TargetTest, execOf(cleanup), mock.Anything).
		Return(failingReport(1, "rm: cannot remove '/tmp/scratch': Permission denied"), nil).Once()

	log := &syncWriter{}
	err := u.Execute(context.Background(), log, j, "worker-1")

	assert.NoError(t, err)
	assert.True(t, j.Status().IsJudged(), "after_* failures are tolerated; got %s", j.Status().String())
	assert.Equal(t, 100.0, j.CalculateScore())
	assert.Contains(t, log.String(), "Exit code: 1")
	assert.Contains(t, log.String(), "Permission denied")
}

// before_each failing mid-run poisons the environment for everything after
// it, so the whole judgment is ERROR — a partial score computed on a broken
// environment would be indistinguishable from a real one.
func TestExecute_BeforeEachNonZero_FailsWholeJudgment(t *testing.T) {
	u, d := newUsecase(t)
	const prep = "sh -c 'exit 3'"

	// Mid-run shape: testcase 1 has no setup and passes; testcase 2's
	// before_each fails. Testcase 3 is never reached.
	tc1 := testcase.NewTestcase(&testcase.TestcaseConfig{ID: 1, Command: "echo one", Score: 30})
	tc1.SetLifecycle(lifecycle.NewTestcaseLifecycle(nil))
	tc2 := testcase.NewTestcase(&testcase.TestcaseConfig{ID: 2, Command: "echo two", Score: 30})
	tc2.SetLifecycle(lifecycle.NewTestcaseLifecycle([]*lifecycle.Lifecycle{lifecycleOf("before_each", prep)}))
	tc3 := testcase.NewTestcase(&testcase.TestcaseConfig{ID: 3, Command: "echo three", Score: 40})
	tc3.SetLifecycle(lifecycle.NewTestcaseLifecycle(nil))
	gi := gradingitem.NewGradingItem(&gradingitem.GradingItemConfig{
		ID: 1, Score: 100, ScoringStrategy: scoringstrategy.IncreaseScoreStrategy, Testcases: []*testcase.Testcase{tc1, tc2, tc3},
	})
	j := judgment.New(&judgment.JudgmentConfig{
		ContestProjectID: 1, SubmissionID: "sub-1",
		Project:      judgmentproject.NewProject(&judgmentproject.ProjectConfig{ID: 1}),
		GradingItems: []*gradingitem.GradingItem{gi},
		Lifecycle:    lifecycle.NewJudgmentLifecycle(nil),
		JudgeType:    judgetype.JUDGE,
	})
	j.SetID("judgment-1")

	expectRunSetup(d)
	d.judgmentRepo.On("SaveProgressIfJudging", mock.Anything).Return(true, nil).Maybe()
	d.environment.On("Exec", mock.Anything, domainservice.TargetTest, execOf("echo one"), mock.Anything).Return(passingReport(), nil).Once()
	d.environment.On("Exec", mock.Anything, domainservice.TargetTest, execOf(prep), mock.Anything).
		Return(failingReport(3, "setup failed"), nil).Once()
	// No Exec for "echo two" / "echo three": the abort must happen first.

	err := u.Execute(context.Background(), &syncWriter{}, j, "worker-1")

	assert.Error(t, err)
	assert.True(t, j.Status().IsError(), "expected ERROR, got %s", j.Status().String())
	if assert.NotNil(t, j.Reason()) {
		assert.Equal(t, executejudgment.LifecycleSetupFailedReason, *j.Reason())
	}
	if assert.NotNil(t, j.ErrorMessage()) {
		assert.Contains(t, *j.ErrorMessage(), "before_each")
	}
	// The completed result is kept; the one that was being set up and the
	// one never reached are swept to Cancelled — nothing is left looking
	// like it is still about to run.
	assert.True(t, tc1.Status().IsPassed(), "completed testcase keeps its result, got %s", tc1.Status().String())
	assert.True(t, tc2.Status().IsCancelled(), "testcase whose setup failed is Cancelled, got %s", tc2.Status().String())
	assert.True(t, tc3.Status().IsCancelled(), "never-reached testcase is Cancelled, got %s", tc3.Status().String())
}
