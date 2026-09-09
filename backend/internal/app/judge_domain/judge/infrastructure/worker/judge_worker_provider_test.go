package worker

import (
	"sync"
	"testing"
	"time"

	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/judgetype"
	"judge-backend/internal/app/judge_domain/judge/domain/queue"
	infraqueue "judge-backend/internal/app/judge_domain/judge/infrastructure/queue"
	judgemocks "judge-backend/internal/app/judge_domain/judge/mocks"
	sharedeventbus "judge-backend/internal/app/shared/infrastructure/eventbus"
	sharedlogger "judge-backend/internal/app/shared/infrastructure/logger"
	"judge-backend/internal/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func newQueueProvider(t *testing.T) queue.JudgeQueueProvider {
	logger := sharedlogger.NewMockLogger()
	return infraqueue.NewJudgeQueueMap(logger)
}

func newTestPool(t *testing.T, provisioner *judgemocks.ExecutionEnvironmentProvisioner) *JudgeWorkerPool {
	logger := sharedlogger.NewMockLogger()
	eb := sharedeventbus.NewEventBus(logger)
	cfg := &config.Config{Judge: &config.JudgeConfig{MaxJudgeDurationMinutes: 30, WorkerCount: 4}}
	return NewJudgeWorkerPool(eb, newQueueProvider(t), nil, provisioner, cfg, logger)
}

func stopAllWorkers(pool *JudgeWorkerPool) {
	for _, w := range pool.ListJudgeWorkers() {
		w.Stop()
	}
}

func TestNewJudgeWorkerPool_CreatesEmptyPool(t *testing.T) {
	pool := newTestPool(t, judgemocks.NewExecutionEnvironmentProvisioner(t))
	assert.Empty(t, pool.ListJudgeWorkers())
}

func TestJudgeWorkerPool_AddJudgeWorker_AddsWorker(t *testing.T) {
	pool := newTestPool(t, judgemocks.NewExecutionEnvironmentProvisioner(t))
	defer stopAllWorkers(pool)

	err := pool.AddJudgeWorker([]judgetype.JudgeType{judgetype.JUDGE})
	assert.NoError(t, err)
	assert.Len(t, pool.ListJudgeWorkers(), 1)
	assert.Equal(t, []judgetype.JudgeType{judgetype.JUDGE}, pool.ListJudgeWorkers()[0].JudgeOrder())
}

func TestInitJudgeWorkerPool_CreatesConfiguredNumberOfWorkers(t *testing.T) {
	pool := newTestPool(t, judgemocks.NewExecutionEnvironmentProvisioner(t))
	defer stopAllWorkers(pool)
	cfg := &config.Config{Judge: &config.JudgeConfig{MaxJudgeDurationMinutes: 30, WorkerCount: 4}}

	InitJudgeWorkerPool(pool, cfg)

	workers := pool.ListJudgeWorkers()
	assert.Len(t, workers, 4)
	// the first three workers each prioritize a different judge type
	assert.Equal(t, judgetype.JUDGE, workers[0].JudgeOrder()[0])
	assert.Equal(t, judgetype.JUDGE_PARTIAL, workers[1].JudgeOrder()[0])
	assert.Equal(t, judgetype.REJUDGE, workers[2].JudgeOrder()[0])
	assert.Equal(t, judgetype.JUDGE, workers[3].JudgeOrder()[0])
}

func TestJudgeWorkerPool_DeleteJudgeWorkerByID_ReleasesRuntimeResources(t *testing.T) {
	provisioner := judgemocks.NewExecutionEnvironmentProvisioner(t)
	pool := newTestPool(t, provisioner)

	assert.NoError(t, pool.AddJudgeWorker([]judgetype.JudgeType{judgetype.JUDGE}))
	workerID := pool.ListJudgeWorkers()[0].ID()
	provisioner.On("ReleaseWorker", workerID).Once()

	pool.DeleteJudgeWorkerByID(workerID)
	assert.Empty(t, pool.ListJudgeWorkers())
}

func TestJudgeWorkerPool_WaitForIdle_ReturnsTrueWhenNoWorkerRunning(t *testing.T) {
	pool := newTestPool(t, judgemocks.NewExecutionEnvironmentProvisioner(t))
	defer stopAllWorkers(pool)

	assert.NoError(t, pool.AddJudgeWorker([]judgetype.JudgeType{judgetype.JUDGE}))
	assert.True(t, pool.WaitForIdle(2*time.Second))
}

// TestJudgeWorkerPool_ConcurrentMembershipOps_NoDataRace is a regression
// test: AddJudgeWorker/DeleteJudgeWorkerByID/ListJudgeWorkers/
// GetJudgeWorkerByID/WaitForIdle used to read and mutate the workers slice
// with no synchronization at all — a concurrent append (which can reallocate
// the backing array) racing a range iteration is undefined behavior, not
// just a benign warning. These methods are genuinely called concurrently in
// production: separate HTTP requests (add/delete/list/get worker), the
// orphan repair services, and the BDD reset seam's WaitForIdle. Meant to be
// run with `go test -race`.
func TestJudgeWorkerPool_ConcurrentMembershipOps_NoDataRace(t *testing.T) {
	provisioner := judgemocks.NewExecutionEnvironmentProvisioner(t)
	provisioner.On("ReleaseWorker", mock.Anything).Maybe()
	pool := newTestPool(t, provisioner)

	const rounds = 50
	var wg sync.WaitGroup

	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for r := 0; r < rounds; r++ {
				assert.NoError(t, pool.AddJudgeWorker([]judgetype.JudgeType{judgetype.JUDGE}))
			}
		}()
	}

	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for r := 0; r < rounds; r++ {
				if workers := pool.ListJudgeWorkers(); len(workers) > 0 {
					pool.DeleteJudgeWorkerByID(workers[0].ID())
				}
			}
		}()
	}

	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for r := 0; r < rounds; r++ {
				if workers := pool.ListJudgeWorkers(); len(workers) > 0 {
					pool.GetJudgeWorkerByID(workers[0].ID())
				}
				pool.WaitForIdle(10 * time.Millisecond)
			}
		}()
	}

	wg.Wait()
	stopAllWorkers(pool)
}
