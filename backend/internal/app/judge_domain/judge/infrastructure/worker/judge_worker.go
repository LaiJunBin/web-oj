package worker

import (
	"context"
	"fmt"
	"io"
	"judge-backend/internal/app/judge_domain/events/judgeworkerlogadded"
	"judge-backend/internal/app/judge_domain/events/judgeworkerstatuschanged"
	executejudgment "judge-backend/internal/app/judge_domain/judge/application/usecase/execute_judgment"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/judgetype"
	"judge-backend/internal/app/judge_domain/judge/domain/queue"
	"judge-backend/internal/app/shared/infrastructure/metrics"
	sharedtypes "judge-backend/internal/app/shared/types"
	"sync"
	"time"
)

type status struct {
	value string
}

var (
	statusInit     = &status{"INIT"}
	statusIdle     = &status{"IDLE"}
	statusStarting = &status{"STARTING"}
	statusRunning  = &status{"RUNNING"}
	statusStopped  = &status{"STOPPED"}
	statusError    = &status{"ERROR"}
)

func (s *status) String() string {
	return s.value
}

type Log struct {
	Text string
	Time time.Time
}

// JudgeWorker pulls judgments from the queues in its judge-order and executes
// them synchronously through the ExecuteJudgment usecase. It owns no runtime
// resources itself — networks and containers belong to the runtime behind the
// usecase.
type JudgeWorker struct {
	id     string
	lock   *sync.Mutex
	mu     sync.RWMutex // guards status, currentJudgeID, and everything else listed below
	logsMu sync.RWMutex
	lifeMu sync.Mutex // serializes Start/Stop/markDeletedAndStop against each other
	reader io.ReadCloser
	writer io.Writer
	logs   []*Log
	status *status

	eventbus         sharedtypes.EventBus
	queueProvider    queue.JudgeQueueProvider
	judgeOrder       []judgetype.JudgeType
	executor         executejudgment.Usecase
	maxJudgeDuration time.Duration
	logger           sharedtypes.Logger

	// lifecycleCtx/lifecycleCancel/done/readerDone belong to the CURRENT
	// generation: Init sets all four exactly once per generation, and
	// nothing renews them mid-generation. processJudges and startLogReader
	// capture ctx/done/readerDone/reader/writer as local parameters (not
	// read from the struct) specifically so a later Init call can never
	// change what a running generation is watching, or racily reassign the
	// io.Pipe out from under a goroutine still reading/writing the old one.
	// lifecycleCtx/lifecycleCancel are deliberately separate from
	// currentJudgeCancel (which cancels only the in-flight judgment) — see
	// CancelJudgeIfCurrent's doc comment for why conflating the two caused
	// a real bug.
	lifecycleCtx    context.Context
	lifecycleCancel context.CancelFunc
	done            chan struct{}
	readerDone      chan struct{}

	// currentJudgeCancel cancels only the in-flight judgment's timeout
	// context; set by processJudges atomically with currentJudgeID (so a
	// caller that sees the ID also sees a valid cancel), cleared by
	// runJudgment, read by CancelJudgeIfCurrent.
	currentJudgeCancel context.CancelFunc

	// deleted is a permanent tombstone set by markDeletedAndStop (called
	// only from JudgeWorkerPool.DeleteJudgeWorkerByID). Once true, Start can
	// never revive this worker again.
	deleted bool

	currentJudgeID string
	// attemptingPop is true from just before a TryPop call until the popped
	// judgment (if any) is claimed via currentJudgeID, or the call returns
	// nothing. It exists so IsBusy can see this worker as busy during that
	// window: TryPop itself is atomic against queue Reset, but claiming
	// currentJudgeID afterward is a second, separate statement, and
	// WaitForIdle samples worker state from another goroutine with no
	// ordering guarantee relative to it otherwise.
	attemptingPop   bool
	createdAt       time.Time
	lastExecutionAt *time.Time
}

// judgeDurationFor picks the judgment's duration budget: the contest
// project's override when the enqueue handler resolved one, the worker's
// configured global default otherwise. Non-positive overrides are ignored —
// a zero/negative timeout would cancel the judgment instantly.
func judgeDurationFor(j *judgment.Judgment, fallback time.Duration) time.Duration {
	if m := j.MaxJudgeDurationMinutes(); m != nil && *m > 0 {
		return time.Duration(*m) * time.Minute
	}
	return fallback
}

func NewJudgeWorker(
	id string,
	lock *sync.Mutex,
	eventbus sharedtypes.EventBus,
	queueProvider queue.JudgeQueueProvider,
	judgeOrder []judgetype.JudgeType,
	executor executejudgment.Usecase,
	maxJudgeDuration time.Duration,
	logger sharedtypes.Logger,
) *JudgeWorker {
	worker := &JudgeWorker{
		id:               id,
		lock:             lock,
		eventbus:         eventbus,
		queueProvider:    queueProvider,
		judgeOrder:       judgeOrder,
		executor:         executor,
		maxJudgeDuration: maxJudgeDuration,
		status:           statusInit,
		currentJudgeID:   "",
		createdAt:        time.Now(),
		logger:           logger,
	}

	worker.Init()

	return worker
}

func (w *JudgeWorker) ID() string {
	return w.id
}

func (w *JudgeWorker) Status() string {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.status.String()
}

func (w *JudgeWorker) SetStatus(s *status) {
	w.mu.Lock()
	w.status = s
	w.mu.Unlock()
	w.logger.Debugf("[JudgeWorker] Worker %s status changed to %s", w.id, s.String())
	w.eventbus.Publish(judgeworkerstatuschanged.NewJudgeWorkerStatusChangedEvent(w.id, s.String()))
}

// transitionStatus atomically moves status from one of `from` to `to`,
// returning whether the transition applied. If the current status isn't in
// `from` — someone else already moved it (most commonly Stop, landing
// between an earlier read of the status and this call) — nothing changes,
// and the caller must not proceed as if its intended transition had taken
// effect. Plain SetStatus is a blind write with no such guard: several
// call sites used to read status, decide to transition based on that read,
// and then call SetStatus unconditionally — a real (non-data-race, so
// -race doesn't catch it) logic race where a concurrent Stop's STOPPED
// could be clobbered back to IDLE or RUNNING by a transition that was
// already in flight before Stop ran.
func (w *JudgeWorker) transitionStatus(from []*status, to *status) bool {
	w.mu.Lock()
	applied := false
	for _, f := range from {
		if w.status == f {
			w.status = to
			applied = true
			break
		}
	}
	w.mu.Unlock()

	if !applied {
		return false
	}
	w.logger.Debugf("[JudgeWorker] Worker %s status changed to %s", w.id, to.String())
	w.eventbus.Publish(judgeworkerstatuschanged.NewJudgeWorkerStatusChangedEvent(w.id, to.String()))
	return true
}

func (w *JudgeWorker) Logs() []*Log {
	w.logsMu.RLock()
	defer w.logsMu.RUnlock()

	logs := make([]*Log, len(w.logs))
	copy(logs, w.logs)
	return logs
}

const maxLogs = 500

func (w *JudgeWorker) AppendLog(log string) {
	l := &Log{
		Text: log,
		Time: time.Now(),
	}
	w.logsMu.Lock()
	w.logs = append(w.logs, l)
	if len(w.logs) > maxLogs {
		w.logs = w.logs[len(w.logs)-maxLogs:]
	}
	w.logsMu.Unlock()
	w.logger.Debugf("[JudgeWorker] Worker %s log: %s", w.id, l.Text)
	w.eventbus.Publish(judgeworkerlogadded.NewJudgeWorkerLogAddedEvent(w.id, l.Text, l.Time))
}

// Reader is guarded because initIO (called from Init, on a Start) reassigns
// w.reader from a different goroutine than callers of this getter — e.g.
// JudgeWorkerPool.DeleteJudgeWorkerByID reads it well outside this worker's
// own lifeMu.
func (w *JudgeWorker) Reader() io.ReadCloser {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.reader
}

func (w *JudgeWorker) JudgeOrder() []judgetype.JudgeType {
	return w.judgeOrder
}

func (w *JudgeWorker) CreatedAt() time.Time {
	return w.createdAt
}

// LastExecutionAt is written by the worker's own processing goroutine
// (setLastExecutionAt, on every judgment pickup) and read from other
// goroutines (worker list/detail query handlers), so both sides go through
// w.mu. The *time.Time value itself is never mutated after creation — each
// call allocates a fresh one — so copying the pointer under the lock is
// sufficient; no need to copy what it points to.
func (w *JudgeWorker) LastExecutionAt() *time.Time {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.lastExecutionAt
}

func (w *JudgeWorker) setLastExecutionAt(t time.Time) {
	w.mu.Lock()
	w.lastExecutionAt = &t
	w.mu.Unlock()
}

// CancelJudgeIfCurrent cancels the in-flight judgment's context, but only if
// it is still the judgment identified by expectedJudgmentID — checked and
// cancelled atomically under one lock, not as two separate calls. Callers
// (the judge_cancelled handler) used to check CurrentJudgeID() and then call
// a parameterless CancelJudge() as two independent steps: if the worker
// finished that judgment and picked up a NEW one in between — entirely
// possible, since finishing and re-claiming happen on the worker's own
// goroutine with no coordination with this caller — the second call would
// cancel whatever judgment happened to be current at that moment, not the
// one the caller actually meant to cancel. Unlike Stop, this does NOT touch
// the worker's lifecycle context: processJudges keeps running on the same
// generation and picks up its next judgment normally afterward. Cancelling
// and renewing the SAME context worker liveness relies on used to make Stop
// unable to reliably tell whether a running processJudges generation had
// actually exited — that's why per-judgment cancellation and
// worker-lifecycle cancellation are two entirely separate context trees.
func (w *JudgeWorker) CancelJudgeIfCurrent(expectedJudgmentID string) bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	if w.currentJudgeID != expectedJudgmentID {
		return false
	}
	if w.currentJudgeCancel != nil {
		w.currentJudgeCancel()
	}
	return true
}

// lifecycleContext returns the current generation's lifecycle context.
// Guarded because Init replaces it from another goroutine (Start).
func (w *JudgeWorker) lifecycleContext() context.Context {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.lifecycleCtx
}

func (w *JudgeWorker) CurrentJudgeID() string {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.currentJudgeID
}

// IsBusy reports whether this worker currently holds a judgment, is running
// one, or is in the brief pop-then-claim window between them. WaitForIdle
// uses this instead of checking CurrentJudgeID/Status directly.
func (w *JudgeWorker) IsBusy() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.currentJudgeID != "" || w.attemptingPop || w.status == statusRunning
}

// initIO creates a fresh pipe for this generation, stores it on the struct
// (for the external Reader() getter and stopLocked's Close call) under
// w.mu, and returns the same reader/writer so the caller can hand them to
// this generation's goroutines as local parameters instead of letting those
// goroutines read the mutable struct fields directly.
//
// The reader is the raw *io.PipeReader, not wrapped in io.NopCloser: that
// wrapping used to make stopLocked's Close call a complete no-op (NopCloser's
// Close never touches the underlying reader), so the log-reader goroutine's
// blocked Read never actually unblocked on Stop — it leaked forever, once
// per Stop→Start cycle.
func (w *JudgeWorker) initIO() (io.ReadCloser, io.Writer) {
	pr, pw := io.Pipe()
	var reader io.ReadCloser = pr
	var writer io.Writer = pw

	w.mu.Lock()
	w.reader = reader
	w.writer = writer
	w.mu.Unlock()

	return reader, writer
}

// processJudges is one generation's worker loop. ctx, done, and writer are
// captured once, as local parameters, at the moment this goroutine was
// spawned by Init — never re-read from the struct — so a later Init (a
// Stop→Start cycle) can't change what THIS goroutine is watching or racily
// swap the pipe out from under it. done is closed unconditionally on exit so
// Stop can wait for this exact generation to be fully gone before returning.
func (w *JudgeWorker) processJudges(ctx context.Context, done chan struct{}, writer io.Writer) {
	defer close(done)
	w.logger.Debugf("[JudgeWorker] Worker %s started processing judges", w.id)
outer:
	for {
		if err := ctx.Err(); err != nil {
			if w.Status() == statusStopped.String() {
				w.logger.Debugf("[JudgeWorker] Worker %s stopped processing judges", w.id)
			} else {
				w.logger.Errorf("[JudgeWorker] Worker %s encountered context error: %v", w.id, err)
				w.SetStatus(statusError)
			}
			break
		}

		currentStatus := w.Status()
		if currentStatus == statusStopped.String() || currentStatus == statusError.String() {
			w.logger.Debugf("[JudgeWorker] Worker %s stopped or encountered error (status: %s)", w.id, currentStatus)
			break
		}
		// Transition INIT/STARTING to IDLE here, gated on the status check
		// above having already passed — not as an unconditional statement
		// before this loop starts. Init sets INIT synchronously before
		// spawning this goroutine, but Go doesn't guarantee this goroutine
		// actually runs before a concurrent Stop (racing to cancel ctx and
		// set STOPPED) does. Using transitionStatus (not a plain SetStatus)
		// closes that regardless of scheduling: if Stop already moved status
		// to STOPPED by the time this runs, the compare-and-set simply
		// doesn't apply, instead of blindly overwriting STOPPED back to
		// IDLE — which would then make this same loop see the (already
		// cancelled) ctx.Err() paired with a status it had just clobbered to
		// non-Stopped, misreporting a clean Stop as an ERROR.
		if currentStatus == statusInit.String() || currentStatus == statusStarting.String() {
			w.transitionStatus([]*status{statusInit, statusStarting}, statusIdle)
			continue
		}
		if currentStatus != statusIdle.String() {
			w.logger.Debugf("[JudgeWorker] Worker %s is not idle (current status: %s), waiting...", w.id, currentStatus)
			sleepOrDone(ctx, 1*time.Second)
			continue
		}

		for _, judgeType := range w.judgeOrder {
			// The whole claim — confirm Idle, pop, set currentJudgeID/
			// currentJudgeCancel, transition to Running — is one ownership
			// boundary, held under w.mu continuously (including across the
			// TryPop call — TryPop only touches the queue map's own
			// separate lock via w.lock, so this can't deadlock with
			// anything). NOT lifeMu: Stop holds lifeMu across its ENTIRE
			// sequence, including its blocking wait on <-done for this very
			// goroutine to exit — if this claim also needed lifeMu, a Stop
			// in progress would hold lifeMu waiting for processJudges to
			// exit, while processJudges sat here waiting for lifeMu to
			// become free to even reach the point where it could exit. w.mu
			// has no such problem: Stop only ever holds it for brief reads/
			// writes, never across its wait, so at worst this claim briefly
			// delays one of those reads/writes, never the other way around.
			//
			// Without holding w.mu across the whole thing, the status check
			// and the claim were two separate moments a concurrent Stop
			// could land between: pop a judgment, have Stop set STOPPED and
			// cancel the lifecycle context, then unconditionally overwrite
			// STOPPED back to RUNNING here anyway — running an
			// already-cancelled judgment and, worse, leaving the worker's
			// final reported status wrong (IDLE or ERROR instead of
			// STOPPED) once everything settles.
			w.mu.Lock()
			if w.status != statusIdle {
				// Stale by the time we got the lock — Stop already moved us.
				// Re-check everything fresh from the top rather than
				// assuming why.
				w.mu.Unlock()
				continue outer
			}
			// attemptingPop is set before TryPop and only cleared once either
			// the judgment is claimed via currentJudgeID or the pop came back
			// empty, so a concurrent IsBusy check can never see a gap between
			// "popped" and "claimed".
			w.attemptingPop = true

			w.lock.Lock()
			judge, err := w.queueProvider.TryPop(judgeType.String())
			w.lock.Unlock()

			if err != nil {
				w.attemptingPop = false
				w.mu.Unlock()
				w.logger.Errorf("[JudgeWorker] Worker %s failed to pop from queue for judge type %s: %v", w.id, judgeType.String(), err)
				// transitionStatus, not SetStatus: the lock was released
				// above, so a concurrent Stop can land in the gap — a blind
				// write here would clobber its STOPPED back to ERROR. Idle is
				// the only status this path can validly be
				// leaving from, since the claim above checked it under the
				// same lock.
				w.transitionStatus([]*status{statusIdle}, statusError)
				continue
			}
			if judge == nil {
				w.attemptingPop = false
				w.mu.Unlock()
				continue
			}

			// runCtx/cancelRun are created here, before currentJudgeID is
			// published, and currentJudgeID/currentJudgeCancel/status are
			// all set together under the same lock — not currentJudgeID
			// here and currentJudgeCancel later inside runJudgment (or
			// status separately still later). Setting them separately left
			// a window where CurrentJudgeID() already reported this
			// judgment as current but currentJudgeCancel was still nil (or
			// still the PREVIOUS judgment's, already cleared); a
			// CancelJudge() landing in that window would read a nil cancel
			// and silently do nothing, even though the judgment it was
			// asked to cancel was already visible as "current".
			runCtx, cancelRun := context.WithTimeout(ctx, judgeDurationFor(judge, w.maxJudgeDuration))
			w.currentJudgeID = judge.ID()
			w.currentJudgeCancel = cancelRun
			w.attemptingPop = false
			w.status = statusRunning
			w.mu.Unlock()
			w.logger.Debugf("[JudgeWorker] Worker %s status changed to %s", w.id, statusRunning.String())
			w.eventbus.Publish(judgeworkerstatuschanged.NewJudgeWorkerStatusChangedEvent(w.id, statusRunning.String()))

			w.runJudgment(runCtx, cancelRun, writer, judgeType.String(), judge)
			// Back to the TOP of the outer loop immediately — re-checking
			// ctx/status — rather than falling through to try the next
			// judgeType in this pass. Continuing the inner loop here used
			// to mean a Stop() that landed while this judgment was running
			// wouldn't be noticed until the whole judgeOrder rotation was
			// exhausted, during which this generation could pop and run
			// MORE judgments after already being told to stop.
			continue outer
		}
		sleepOrDone(ctx, 1*time.Second)
	}
}

// sleepOrDone waits for d or until ctx is cancelled, whichever comes first.
// processJudges uses this instead of a blind time.Sleep so Stop doesn't have
// to wait out the full poll interval on top of waiting for this generation
// to notice it should exit.
func sleepOrDone(ctx context.Context, d time.Duration) {
	select {
	case <-ctx.Done():
	case <-time.After(d):
	}
}

// runJudgment executes one judgment synchronously. runCtx/cancelRun are
// built by the caller (processJudges) BEFORE currentJudgeID is published, so
// that currentJudgeID and currentJudgeCancel become visible together — see
// the comment at the call site for why. writer is this generation's pipe
// writer, passed down rather than read from w.writer directly — see
// initIO's doc comment.
func (w *JudgeWorker) runJudgment(runCtx context.Context, cancelRun context.CancelFunc, writer io.Writer, queueName string, judge *judgment.Judgment) {
	w.logger.Infof("[JudgeWorker] Worker %s picked judge %s from queue %s", w.id, judge.ID(), queueName)
	w.AppendLog(fmt.Sprintf("Pick judge %s from queue %s\n", judge.ID(), queueName))

	// Status is already Running by this point — set as part of the atomic
	// claim in processJudges, not here; see that call site's comment.
	w.setLastExecutionAt(time.Now())

	metrics.IncActiveJudgments()
	executionStart := time.Now()
	if err := w.executor.Execute(runCtx, writer, judge, w.id); err != nil {
		w.logger.Errorf("[JudgeWorker] Worker %s finished judge %s with error: %v", w.id, judge.ID(), err)
	}
	metrics.DecActiveJudgments()
	// Execute mutated judge to its terminal status, whatever it was.
	metrics.ObserveJudgmentFinished(judge.Status().String(), time.Since(executionStart).Seconds())
	cancelRun()

	w.mu.Lock()
	w.currentJudgeID = ""
	w.currentJudgeCancel = nil
	w.mu.Unlock()
	w.AppendLog("--- Judge finished ---")
	w.logger.Infof("[JudgeWorker] Worker %s finished judge %s", w.id, judge.ID())
	// transitionStatus, not a read-then-SetStatus: a plain "if RUNNING then
	// SetStatus(Idle)" reads and writes as two separate steps, so a Stop
	// landing in between (setting STOPPED) would get clobbered back to IDLE
	// by this call proceeding on its stale read.
	w.transitionStatus([]*status{statusRunning}, statusIdle)
}

// startLogReader is one generation's log-forwarding loop. reader and
// readerDone are captured once, as local parameters, for the same reason
// processJudges captures ctx/done — see initIO's doc comment. readerDone is
// closed unconditionally on exit so Stop can wait for this goroutine too,
// not just processJudges.
func (w *JudgeWorker) startLogReader(reader io.ReadCloser, readerDone chan struct{}) {
	defer close(readerDone)
	w.logger.Debugf("[JudgeWorker] Worker %s started log reader", w.id)
	for {
		buf := make([]byte, 1024)
		n, err := reader.Read(buf)
		if err != nil {
			if err == io.EOF {
				w.logger.Debugf("[JudgeWorker] Worker %s log reader reached EOF", w.id)
				break
			}
			if w.Status() != statusStopped.String() && w.lifecycleContext().Err() == nil {
				w.SetStatus(statusError)
			}
			w.logger.Errorf("[JudgeWorker] Worker %s log reader encountered error: %v", w.id, err)
			break
		}
		w.AppendLog(string(buf[:n]))
	}
}

func (w *JudgeWorker) Init() {
	w.logger.Debugf("[JudgeWorker] Initializing worker %s", w.id)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	readerDone := make(chan struct{})
	reader, writer := w.initIO()

	w.mu.Lock()
	w.lifecycleCtx = ctx
	w.lifecycleCancel = cancel
	w.done = done
	w.readerDone = readerDone
	w.currentJudgeID = ""
	w.mu.Unlock()

	w.SetStatus(statusInit)
	w.AppendLog("\n-----init-----\n")

	go w.processJudges(ctx, done, writer)
	go w.startLogReader(reader, readerDone)
	w.logger.Debugf("[JudgeWorker] Worker %s initialization completed", w.id)
}

func (w *JudgeWorker) Stop() {
	w.lifeMu.Lock()
	defer w.lifeMu.Unlock()
	w.stopLocked()
}

// stopLocked runs the actual stop sequence. Callers must already hold
// lifeMu (Stop and markDeletedAndStop are the only two).
func (w *JudgeWorker) stopLocked() {
	if w.Status() == statusStopped.String() {
		w.logger.Debugf("[JudgeWorker] Worker %s already stopped", w.id)
		return
	}

	w.mu.RLock()
	cancel := w.lifecycleCancel
	done := w.done
	readerDone := w.readerDone
	reader := w.reader
	w.mu.RUnlock()

	// Status set BEFORE cancelling, not after: processJudges treats
	// "context cancelled but status isn't Stopped yet" as an unexpected
	// error, not a normal shutdown. Cancelling first would leave exactly
	// that window open — processJudges could observe the cancelled context
	// on its very next loop check before this goroutine got around to
	// setting the status, misclassifying a clean Stop as an ERROR.
	w.SetStatus(statusStopped)
	if cancel != nil {
		cancel()
	}
	if reader != nil {
		reader.Close()
	}
	w.AppendLog("\n-----stop-----\n")

	// Wait for THIS generation's processJudges AND log reader to fully exit
	// before returning — not just processJudges. Without waiting for
	// processJudges: a judgment still mid-execution when Stop is called
	// could keep the generation running well past this call, and a Start()
	// that follows immediately would spin up a SECOND processJudges
	// goroutine via Init while the first was still alive, both consuming
	// the same queues at once. The log reader no longer risks that specific
	// double-consumer failure mode (it only touches this generation's own
	// local reader value, never the mutable struct fields), but waiting for
	// it too means Stop's contract stays simple and absolute: when it
	// returns, nothing from this generation is still running.
	if done != nil {
		<-done
	}
	if readerDone != nil {
		<-readerDone
	}
	w.logger.Infof("[JudgeWorker] Worker %s stopped", w.id)
}

// markDeletedAndStop permanently tombstones the worker — Start can never
// revive it again, even a Start call already blocked waiting on lifeMu —
// and stops it, both under the SAME lifeMu critical section Start's own
// check runs in. Called only by JudgeWorkerPool.DeleteJudgeWorkerByID, after
// the worker has already been removed from the pool's slice: without this,
// a Start call made through a handle obtained before the delete (e.g. via
// GetJudgeWorkerByID) could still revive a worker the pool has already
// committed to removing — reviving an orphan that ListJudgeWorkers/
// GetJudgeWorkerByID/DeleteJudgeWorkerByID can never see or reach again,
// after ReleaseWorker has already run for it.
func (w *JudgeWorker) markDeletedAndStop() {
	w.lifeMu.Lock()
	defer w.lifeMu.Unlock()

	w.mu.Lock()
	w.deleted = true
	w.mu.Unlock()

	w.stopLocked()
}

func (w *JudgeWorker) Start() {
	w.lifeMu.Lock()
	defer w.lifeMu.Unlock()

	w.mu.RLock()
	deleted := w.deleted
	w.mu.RUnlock()
	if deleted {
		w.logger.Debugf("[JudgeWorker] Worker %s start skipped: worker has been deleted", w.id)
		return
	}

	currentStatus := w.Status()
	if currentStatus != statusStopped.String() && currentStatus != statusError.String() {
		w.logger.Debugf("[JudgeWorker] Worker %s start skipped because current status is %s", w.id, currentStatus)
		return
	}

	w.logger.Debugf("[JudgeWorker] Starting worker %s", w.id)
	// stopLocked is a no-op if status is already Stopped (the common case —
	// whatever set it already ran this cleanup), but it's essential when
	// recovering from ERROR: ERROR can be set by the log reader or a
	// queue-pop failure while the OLD generation's processJudges is still
	// actively running a judgment, not yet exited — Init below replaces the
	// context/pipe/done channels unconditionally, so without draining the
	// old generation first, this would spin up a SECOND processJudges
	// goroutine while the first was still alive: the exact double-consumer
	// bug Stop's own wait-for-done exists to prevent, just reached through
	// the ERROR path instead of a direct Stop call.
	w.stopLocked()
	w.SetStatus(statusStarting)
	w.AppendLog("\n-----start-----\n")
	w.Init()
	w.logger.Debugf("[JudgeWorker] Worker %s started", w.id)
}
