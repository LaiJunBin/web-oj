package worker

import (
	"context"
	"fmt"
	"io"
	"runtime"
	"sync"
	"testing"
	"time"

	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/judgetype"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/project"
	"judge-backend/internal/app/judge_domain/judge/domain/prioritystrategies"
	infraqueue "judge-backend/internal/app/judge_domain/judge/infrastructure/queue"
	judgemocks "judge-backend/internal/app/judge_domain/judge/mocks"
	sharedeventbus "judge-backend/internal/app/shared/infrastructure/eventbus"
	sharedlogger "judge-backend/internal/app/shared/infrastructure/logger"

	"github.com/stretchr/testify/mock"
)

// noopExecutor is a plain (non-mock) executejudgment.Usecase stub for tests
// that race Stop against a worker's own claim: testify's mock machinery
// internally reflects over call arguments (for its expectation-diff
// bookkeeping) — including the io.Writer argument, which wraps the
// generation's *io.PipeWriter — and that reflection can genuinely race a
// concurrent Stop closing the same pipe's reader half (they share one
// underlying pipe struct). That's an artifact of the mock's own
// introspection, not a production bug: real Execute never reflects over its
// arguments, it only calls Write on them. noopExecutor sidesteps it
// entirely by doing nothing but returning nil.
type noopExecutor struct{}

func (noopExecutor) Execute(_ context.Context, _ io.Writer, _ *judgment.Judgment, _ string) error {
	return nil
}

// TestJudgeWorker_LastExecutionAt_ConcurrentReadWrite_NoDataRace is a
// regression test: setLastExecutionAt (called from the worker's own
// processing goroutine on every judgment pickup) and LastExecutionAt (called
// from other goroutines — the worker list/detail query handlers) used to
// touch w.lastExecutionAt with no synchronization at all. Meant to be run
// with `go test -race`.
func TestJudgeWorker_LastExecutionAt_ConcurrentReadWrite_NoDataRace(t *testing.T) {
	w := &JudgeWorker{}

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		for i := 0; i < 200; i++ {
			w.setLastExecutionAt(time.Now())
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 200; i++ {
			_ = w.LastExecutionAt()
		}
	}()

	wg.Wait()
}

func newWorkerTestJudgment(id string) *judgment.Judgment {
	j := judgment.New(&judgment.JudgmentConfig{
		Project:   project.NewProject(&project.ProjectConfig{ID: 1}),
		JudgeType: judgetype.JUDGE,
		CreatedAt: time.Now(),
	})
	j.SetID(id)
	return j
}

// TestJudgeWorker_Stop_WaitsForGenerationToFullyExit is a regression test:
// Stop used to only cancel the context and return, without waiting for
// processJudges to actually observe it and exit. A judgment still executing
// at that moment could keep running — and worse, once it finished, the old
// processJudges loop used to fall through to try the NEXT judgeType instead
// of re-checking status, so it could pop and run further judgments after
// Stop had already returned. This proves both halves are closed: Stop
// genuinely blocks until the in-flight judgment's generation is gone, and no
// further judgment is picked up after Stop returns.
func TestJudgeWorker_Stop_WaitsForGenerationToFullyExit(t *testing.T) {
	logger := sharedlogger.NewMockLogger()
	eventbus := sharedeventbus.NewEventBus(logger)
	queueProvider := infraqueue.NewJudgeQueueMap(logger)

	q, err := queueProvider.GetQueue(judgetype.JUDGE.String())
	if err != nil {
		t.Fatalf("GetQueue failed: %v", err)
	}
	// Seed two judgments: the first is what the worker will be executing
	// when Stop is called; the second must still be sitting in the queue,
	// untouched, once Stop returns.
	q.Push(newWorkerTestJudgment("judgment-executing"), prioritystrategies.NewDefault())
	q.Push(newWorkerTestJudgment("judgment-should-not-be-popped"), prioritystrategies.NewDefault())

	executor := judgemocks.NewUsecase(t)
	executing := make(chan struct{})
	unblock := make(chan struct{})
	executor.On("Execute", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Run(func(mock.Arguments) {
			close(executing)
			<-unblock
		}).
		Return(nil).Once()

	w := NewJudgeWorker("worker-1", &sync.Mutex{}, eventbus, queueProvider, []judgetype.JudgeType{judgetype.JUDGE}, executor, time.Minute, logger)

	select {
	case <-executing:
	case <-time.After(2 * time.Second):
		t.Fatal("worker never started executing the seeded judgment")
	}

	stopDone := make(chan struct{})
	go func() {
		defer close(stopDone)
		w.Stop()
	}()

	select {
	case <-stopDone:
		t.Fatal("Stop returned before the in-flight judgment finished")
	case <-time.After(100 * time.Millisecond):
		// expected: Stop is still blocked, waiting on the generation to exit
	}

	close(unblock)

	select {
	case <-stopDone:
	case <-time.After(2 * time.Second):
		t.Fatal("Stop did not return after the in-flight judgment finished")
	}

	if got := w.Status(); got != statusStopped.String() {
		t.Fatalf("expected status Stopped after Stop returned, got %s", got)
	}

	// Give any errant extra pop a moment to happen, then confirm it didn't.
	time.Sleep(50 * time.Millisecond)
	remaining, err := queueProvider.TryPop(judgetype.JUDGE.String())
	if err != nil {
		t.Fatalf("TryPop failed: %v", err)
	}
	if remaining == nil || remaining.ID() != "judgment-should-not-be-popped" {
		t.Fatalf("expected the second judgment to still be queued untouched, got %v", remaining)
	}
}

// TestJudgeWorker_CancelJudge_NeverMissesJudgmentAlreadyVisibleAsCurrent is a
// regression test for a race introduced by the fix above: currentJudgeID
// used to be published (under lock) well before currentJudgeCancel was set
// (inside runJudgment, after a separate lock acquisition). A CancelJudge()
// call landing in that window would read a nil (or stale, already-cleared)
// currentJudgeCancel and silently do nothing — even though CurrentJudgeID()
// already reported the judgment as current, so a caller had every reason to
// believe cancelling it would work. Runs many iterations racing CancelJudge
// against the worker's own claim sequence — busy-polling CurrentJudgeID to
// catch the earliest possible moment — since this is a narrow timing window,
// not a deterministically reproducible ordering.
func TestJudgeWorker_CancelJudge_NeverMissesJudgmentAlreadyVisibleAsCurrent(t *testing.T) {
	for iter := 0; iter < 200; iter++ {
		logger := sharedlogger.NewMockLogger()
		eventbus := sharedeventbus.NewEventBus(logger)
		queueProvider := infraqueue.NewJudgeQueueMap(logger)
		q, err := queueProvider.GetQueue(judgetype.JUDGE.String())
		if err != nil {
			t.Fatalf("GetQueue failed: %v", err)
		}
		q.Push(newWorkerTestJudgment("judgment-1"), prioritystrategies.NewDefault())

		executor := judgemocks.NewUsecase(t)
		cancelled := make(chan struct{})
		executor.On("Execute", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) {
				runCtx := args.Get(0).(context.Context)
				<-runCtx.Done()
				close(cancelled)
			}).
			Return(context.Canceled).Once()

		w := NewJudgeWorker("worker-1", &sync.Mutex{}, eventbus, queueProvider, []judgetype.JudgeType{judgetype.JUDGE}, executor, time.Minute, logger)

		for w.CurrentJudgeID() == "" {
			runtime.Gosched()
		}
		if !w.CancelJudgeIfCurrent("judgment-1") {
			t.Fatalf("iteration %d: CancelJudgeIfCurrent reported judgment-1 as not current, even though CurrentJudgeID() had already reported it", iter)
		}

		select {
		case <-cancelled:
		case <-time.After(2 * time.Second):
			t.Fatalf("iteration %d: CancelJudgeIfCurrent did not cancel the judgment that was already visible as current", iter)
		}

		w.Stop()
	}
}

// TestJudgeWorker_Stop_RacingClaim_FinalStatusAlwaysStopped is a regression
// test: the Idle-confirm -> pop -> claim -> Running transition used to be
// several separate lock acquisitions, so Stop landing partway through could
// get unconditionally overwritten back to Running (or, once processJudges
// looped back around, misreported as Error) instead of the final status
// correctly settling on Stopped. Races Stop against the worker's own very
// first claim attempt (right after construction, when the freshly spawned
// processJudges goroutine hasn't necessarily run yet) across many
// iterations — the exact interleaving isn't controlled, but the invariant
// (final status is always exactly Stopped once Stop returns) must hold
// regardless of how the race resolves.
func TestJudgeWorker_Stop_RacingClaim_FinalStatusAlwaysStopped(t *testing.T) {
	for iter := 0; iter < 200; iter++ {
		logger := sharedlogger.NewMockLogger()
		eventbus := sharedeventbus.NewEventBus(logger)
		queueProvider := infraqueue.NewJudgeQueueMap(logger)

		q, err := queueProvider.GetQueue(judgetype.JUDGE.String())
		if err != nil {
			t.Fatalf("GetQueue failed: %v", err)
		}
		q.Push(newWorkerTestJudgment(fmt.Sprintf("judgment-%d", iter)), prioritystrategies.NewDefault())

		w := NewJudgeWorker("worker-1", &sync.Mutex{}, eventbus, queueProvider, []judgetype.JudgeType{judgetype.JUDGE}, noopExecutor{}, time.Minute, logger)
		w.Stop()

		if got := w.Status(); got != statusStopped.String() {
			t.Fatalf("iteration %d: expected final status Stopped after Stop returned, got %s", iter, got)
		}
	}
}

// TestJudgeWorker_TransitionStatus_DoesNotApplyWhenCurrentStatusDiffers is a
// direct unit test of transitionStatus's compare-and-set contract: several
// call sites (INIT/STARTING->IDLE, RUNNING->IDLE) used to read status and
// then unconditionally SetStatus based on that stale read — a real (though
// not -race-detectable) logic race where a concurrent Stop's STOPPED could
// get clobbered back to IDLE/RUNNING by a transition already in flight.
func TestJudgeWorker_TransitionStatus_DoesNotApplyWhenCurrentStatusDiffers(t *testing.T) {
	logger := sharedlogger.NewMockLogger()
	w := &JudgeWorker{status: statusStopped, logger: logger, eventbus: sharedeventbus.NewEventBus(logger)}

	if applied := w.transitionStatus([]*status{statusRunning}, statusIdle); applied {
		t.Fatal("expected transitionStatus to report not applied when current status doesn't match `from`")
	}
	if got := w.Status(); got != statusStopped.String() {
		t.Fatalf("expected status to remain Stopped, got %s", got)
	}
}

func TestJudgeWorker_TransitionStatus_AppliesWhenCurrentStatusMatches(t *testing.T) {
	logger := sharedlogger.NewMockLogger()
	w := &JudgeWorker{status: statusRunning, logger: logger, eventbus: sharedeventbus.NewEventBus(logger)}

	if applied := w.transitionStatus([]*status{statusRunning}, statusIdle); !applied {
		t.Fatal("expected transitionStatus to apply when current status matches `from`")
	}
	if got := w.Status(); got != statusIdle.String() {
		t.Fatalf("expected status Idle, got %s", got)
	}
}

// TestJudgeWorker_StartFromError_DrainsOldGenerationBeforeReinitializing is a
// regression test: Start used to accept ERROR status and call Init directly,
// with no cancel/drain step — unlike recovering from STOPPED, where
// whatever set that status had already gone through stopLocked. ERROR can be
// set (by the log reader, or a queue-pop failure) while the OLD
// generation's processJudges is still actively running a judgment, not yet
// exited. Init replaces the context/pipe/done channels unconditionally, so
// starting from ERROR without draining first could spin up a SECOND
// processJudges goroutine while the first was still alive — the exact
// double-consumer bug Stop's wait-for-done exists to prevent, just reached
// through ERROR instead of a direct Stop call. The executor mock's Once()
// is the actual assertion here: if the old generation's judgment were
// picked up a second time by a rogue second processJudges goroutine, the
// mock would panic on the unexpected extra call.
func TestJudgeWorker_StartFromError_DrainsOldGenerationBeforeReinitializing(t *testing.T) {
	logger := sharedlogger.NewMockLogger()
	eventbus := sharedeventbus.NewEventBus(logger)
	queueProvider := infraqueue.NewJudgeQueueMap(logger)

	q, err := queueProvider.GetQueue(judgetype.JUDGE.String())
	if err != nil {
		t.Fatalf("GetQueue failed: %v", err)
	}
	q.Push(newWorkerTestJudgment("judgment-executing"), prioritystrategies.NewDefault())

	executor := judgemocks.NewUsecase(t)
	executing := make(chan struct{})
	unblock := make(chan struct{})
	executor.On("Execute", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Run(func(mock.Arguments) {
			close(executing)
			<-unblock
		}).
		Return(nil).Once()

	w := NewJudgeWorker("worker-1", &sync.Mutex{}, eventbus, queueProvider, []judgetype.JudgeType{judgetype.JUDGE}, executor, time.Minute, logger)

	select {
	case <-executing:
	case <-time.After(2 * time.Second):
		t.Fatal("worker never started executing the seeded judgment")
	}

	// Simulate the log reader (or a queue-pop failure) setting ERROR while
	// this generation is still actively running a judgment.
	w.SetStatus(statusError)

	startDone := make(chan struct{})
	go func() {
		defer close(startDone)
		w.Start()
	}()

	select {
	case <-startDone:
		t.Fatal("Start returned before the in-flight judgment from the OLD generation finished")
	case <-time.After(100 * time.Millisecond):
		// expected: Start is blocked inside stopLocked, waiting on the old
		// generation to exit
	}

	close(unblock)

	select {
	case <-startDone:
	case <-time.After(2 * time.Second):
		t.Fatal("Start did not return after the in-flight judgment finished")
	}

	w.Stop()
}

// TestJudgeWorker_Stop_LogReaderGoroutineActuallyExits is a regression test
// for a bug found while reviewing the Stop/Start fix above: initIO used to
// wrap the pipe reader in io.NopCloser before storing it, whose Close is a
// permanent no-op regardless of the underlying reader — so stopLocked's
// Close call never actually closed the pipe, and startLogReader's blocked
// Read never unblocked on Stop. It leaked forever, once per Stop→Start
// cycle. Because Stop now also waits on readerDone (closed when
// startLogReader exits), that bug would manifest here as Stop hanging
// indefinitely — this test's timeout is what catches a regression, not an
// assertion on state.
func TestJudgeWorker_Stop_LogReaderGoroutineActuallyExits(t *testing.T) {
	logger := sharedlogger.NewMockLogger()
	eventbus := sharedeventbus.NewEventBus(logger)
	queueProvider := infraqueue.NewJudgeQueueMap(logger)
	executor := judgemocks.NewUsecase(t) // no calls expected: queue stays empty

	w := NewJudgeWorker("worker-1", &sync.Mutex{}, eventbus, queueProvider, []judgetype.JudgeType{judgetype.JUDGE}, executor, time.Minute, logger)

	stopDone := make(chan struct{})
	go func() {
		defer close(stopDone)
		w.Stop()
	}()

	select {
	case <-stopDone:
	case <-time.After(2 * time.Second):
		t.Fatal("Stop did not return — the log reader goroutine likely never unblocked")
	}
}

// TestJudgeWorker_StartAfterMarkDeletedAndStop_IsNoOp is a regression test
// for the pool-Delete/Start race: once a worker has been tombstoned via
// markDeletedAndStop (what JudgeWorkerPool.DeleteJudgeWorkerByID calls),
// Start must never revive it again — even though its status looks
// identical (Stopped) to a worker that was merely paused via the regular
// reversible Stop.
func TestJudgeWorker_StartAfterMarkDeletedAndStop_IsNoOp(t *testing.T) {
	logger := sharedlogger.NewMockLogger()
	eventbus := sharedeventbus.NewEventBus(logger)
	queueProvider := infraqueue.NewJudgeQueueMap(logger)
	executor := judgemocks.NewUsecase(t) // no calls expected

	w := NewJudgeWorker("worker-1", &sync.Mutex{}, eventbus, queueProvider, []judgetype.JudgeType{judgetype.JUDGE}, executor, time.Minute, logger)

	w.markDeletedAndStop()
	if got := w.Status(); got != statusStopped.String() {
		t.Fatalf("expected status Stopped after markDeletedAndStop, got %s", got)
	}

	w.Start()

	if got := w.Status(); got != statusStopped.String() {
		t.Fatalf("expected Start on a deleted worker to be a no-op (status still Stopped), got %s", got)
	}
}
