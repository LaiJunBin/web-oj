package handler

import (
	"context"
	"testing"
	"time"

	"judge-backend/internal/app/judge_domain/events/judgecancelled"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/judgetype"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/project"
	"judge-backend/internal/app/judge_domain/judge/domain/prioritystrategies"
	infraqueue "judge-backend/internal/app/judge_domain/judge/infrastructure/queue"
	"judge-backend/internal/app/judge_domain/judge/infrastructure/worker"
	judgemocks "judge-backend/internal/app/judge_domain/judge/mocks"
	sharedeventbus "judge-backend/internal/app/shared/infrastructure/eventbus"
	sharedlogger "judge-backend/internal/app/shared/infrastructure/logger"
	sharedmocks "judge-backend/internal/app/shared/mocks"
	"judge-backend/internal/config"

	"github.com/stretchr/testify/mock"
)

func newJudgeCancelledTestJudgment(id string) *judgment.Judgment {
	j := judgment.New(&judgment.JudgmentConfig{
		Project:   project.NewProject(&project.ProjectConfig{ID: 1}),
		JudgeType: judgetype.JUDGE,
		CreatedAt: time.Now(),
	})
	j.SetID(id)
	return j
}

// TestJudgeCancelled_CancelsTheWorkerActuallyHandlingIt exercises the full
// dispatch path with a real JudgeWorkerPool/JudgeWorker (not mocks): seeds a
// judgment whose execution the handler must interrupt, and asserts the
// executor's context is actually cancelled. The atomic
// CancelJudgeIfCurrent guarantee itself is covered exhaustively at the
// JudgeWorker level (see judge_worker_test.go); this test covers the
// handler's own iterate-and-dispatch logic against the real pool type.
func TestJudgeCancelled_CancelsTheWorkerActuallyHandlingIt(t *testing.T) {
	logger := sharedlogger.NewMockLogger()
	eventbus := sharedeventbus.NewEventBus(logger)
	queueProvider := infraqueue.NewJudgeQueueMap(logger)

	q, err := queueProvider.GetQueue(judgetype.JUDGE.String())
	if err != nil {
		t.Fatalf("GetQueue failed: %v", err)
	}
	q.Push(newJudgeCancelledTestJudgment("judgment-1"), prioritystrategies.NewDefault())

	executor := judgemocks.NewUsecase(t)
	executing := make(chan struct{})
	cancelledCh := make(chan struct{})
	executor.On("Execute", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			close(executing)
			runCtx := args.Get(0).(context.Context)
			<-runCtx.Done()
			close(cancelledCh)
		}).
		Return(context.Canceled).Once()

	provisioner := judgemocks.NewExecutionEnvironmentProvisioner(t)
	cfg := &config.Config{Judge: &config.JudgeConfig{MaxJudgeDurationMinutes: 30}}
	pool := worker.NewJudgeWorkerPool(eventbus, queueProvider, executor, provisioner, cfg, logger)
	if err := pool.AddJudgeWorker([]judgetype.JudgeType{judgetype.JUDGE}); err != nil {
		t.Fatalf("AddJudgeWorker failed: %v", err)
	}
	defer func() {
		for _, w := range pool.ListJudgeWorkers() {
			w.Stop()
		}
	}()

	select {
	case <-executing:
	case <-time.After(2 * time.Second):
		t.Fatal("worker never started executing the seeded judgment")
	}

	event := judgecancelled.NewJudgeCancelledEvent("judgment-1")
	h := &JudgeCancelledHandler{
		workerPool: pool,
		logger:     logger,
	}

	h.JudgeCancelled(event)

	select {
	case <-cancelledCh:
	case <-time.After(2 * time.Second):
		t.Fatal("expected the executing judgment's context to be cancelled")
	}
}

// TestJudgeCancelled_NoWorkerHandlingIt_NoOp asserts the handler doesn't
// panic or misbehave when no worker is currently handling the target
// judgment (e.g. it already finished on its own).
func TestJudgeCancelled_NoWorkerHandlingIt_NoOp(t *testing.T) {
	logger := sharedlogger.NewMockLogger()
	eventbus := sharedeventbus.NewEventBus(logger)
	queueProvider := infraqueue.NewJudgeQueueMap(logger)
	executor := judgemocks.NewUsecase(t) // no calls expected: queue stays empty

	provisioner := judgemocks.NewExecutionEnvironmentProvisioner(t)
	cfg := &config.Config{Judge: &config.JudgeConfig{MaxJudgeDurationMinutes: 30}}
	pool := worker.NewJudgeWorkerPool(eventbus, queueProvider, executor, provisioner, cfg, logger)
	if err := pool.AddJudgeWorker([]judgetype.JudgeType{judgetype.JUDGE}); err != nil {
		t.Fatalf("AddJudgeWorker failed: %v", err)
	}
	defer func() {
		for _, w := range pool.ListJudgeWorkers() {
			w.Stop()
		}
	}()

	event := judgecancelled.NewJudgeCancelledEvent("judgment-missing")
	h := &JudgeCancelledHandler{
		workerPool: pool,
		logger:     logger,
	}

	h.JudgeCancelled(event)
}

// TestJudgeCancelled_WiringPin_ClaimProtocolAndKey pins the registration
// (same template as TestCancelJudge_WiringPin_ClaimProtocolAndKey): the
// claimed protocol must be attempted, and the claim key here is the RAW
// judgment ID — JudgeCancelledEvent.ID() returns it directly. That makes
// this the only unprefixed claim keyspace, the one the contest scoreboard
// handler's "judged-contest-" prefix deliberately steers clear of; the exact
// string assertion locks that layout in.
func TestJudgeCancelled_WiringPin_ClaimProtocolAndKey(t *testing.T) {
	logger := sharedlogger.NewMockLogger()
	eb := sharedeventbus.NewEventBus(logger)
	queueProvider := infraqueue.NewJudgeQueueMap(logger)
	// The Register parameter is the concrete *JudgeWorkerPool, so a real one
	// is built — but with no workers added, and the denied claim keeps the
	// handler from ever running, so executor/provisioner stay bare strict
	// mocks with no expectations.
	pool := worker.NewJudgeWorkerPool(eb, queueProvider,
		judgemocks.NewUsecase(t), judgemocks.NewExecutionEnvironmentProvisioner(t),
		&config.Config{Judge: &config.JudgeConfig{MaxJudgeDurationMinutes: 30}}, logger)
	eventRepo := sharedmocks.NewEventRepository(t)
	eventRepo.On("TryClaimEvent", "judgment-1").Return(false, nil).Once()

	RegisterJudgeCancelledHandler(eb, eventRepo, pool, logger)

	eb.Publish(judgecancelled.NewJudgeCancelledEvent("judgment-1"))
	eb.WaitAsync()
}
