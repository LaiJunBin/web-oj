package service

import (
	"testing"

	"judge-backend/internal/app/judge_domain/judge/infrastructure/queue"
	"judge-backend/internal/app/judge_domain/judge/infrastructure/worker"
	judgemocks "judge-backend/internal/app/judge_domain/judge/mocks"
	sharedeventbus "judge-backend/internal/app/shared/infrastructure/eventbus"
	sharedlogger "judge-backend/internal/app/shared/infrastructure/logger"
	"judge-backend/internal/config"

	"github.com/stretchr/testify/mock"
)

func newCleanupWorkerPool(t *testing.T, provisioner *judgemocks.ExecutionEnvironmentProvisioner) *worker.JudgeWorkerPool {
	logger := sharedlogger.NewMockLogger()
	eb := sharedeventbus.NewEventBus(logger)
	queueProvider := queue.NewJudgeQueueMap(logger)
	cfg := &config.Config{Judge: &config.JudgeConfig{MaxJudgeDurationMinutes: 30}}
	// NewJudgeWorkerPool no longer creates workers; InitJudgeWorkerPool does.
	// The cleanup service runs before InitJudgeWorkerPool, so the pool is empty here.
	return worker.NewJudgeWorkerPool(eb, queueProvider, nil, provisioner, cfg, logger)
}

func TestOrphanContainerCleanupService_CleanupOnce_DelegatesToProvisioner(t *testing.T) {
	provisioner := judgemocks.NewExecutionEnvironmentProvisioner(t)
	pool := newCleanupWorkerPool(t, provisioner)
	service := NewOrphanContainerCleanupService(provisioner, pool, sharedlogger.NewMockLogger())

	provisioner.On("CleanupOrphans",
		mock.MatchedBy(func(m map[string]struct{}) bool { return len(m) == 0 }),
		mock.MatchedBy(func(m map[string]struct{}) bool { return len(m) == 0 }),
	).Once()

	service.CleanupOnce()
}
