package worker

import (
	executejudgment "judge-backend/internal/app/judge_domain/judge/application/usecase/execute_judgment"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/judgetype"
	"judge-backend/internal/app/judge_domain/judge/domain/queue"
	domainservice "judge-backend/internal/app/judge_domain/judge/domain/service"
	sharedtypes "judge-backend/internal/app/shared/types"
	"judge-backend/internal/config"
	"sync"
	"time"

	"github.com/google/uuid"
)

type JudgeWorkerPool struct {
	eventbus      sharedtypes.EventBus
	queueProvider queue.JudgeQueueProvider
	executor      executejudgment.Usecase
	provisioner   domainservice.ExecutionEnvironmentProvisioner
	lock          *sync.Mutex
	// workersMu guards workers. Deliberately separate from lock: lock is
	// handed to every JudgeWorker to serialize their queue-pop attempts
	// against each other, an unrelated concern — reusing it here would mean
	// a slow pool-membership operation (or vice versa) blocks the other.
	workersMu        sync.RWMutex
	workers          []*JudgeWorker
	maxJudgeDuration time.Duration
	logger           sharedtypes.Logger
}

func NewJudgeWorkerPool(
	eventbus sharedtypes.EventBus,
	queueProvider queue.JudgeQueueProvider,
	executor executejudgment.Usecase,
	provisioner domainservice.ExecutionEnvironmentProvisioner,
	cfg *config.Config,
	logger sharedtypes.Logger,
) *JudgeWorkerPool {
	pool := &JudgeWorkerPool{
		eventbus:         eventbus,
		queueProvider:    queueProvider,
		executor:         executor,
		provisioner:      provisioner,
		lock:             &sync.Mutex{},
		maxJudgeDuration: time.Duration(cfg.Judge.MaxJudgeDurationMinutes) * time.Minute,
		logger:           logger,
		workers:          []*JudgeWorker{},
	}
	pool.logger.Infof("[JudgeWorkerPool] Created new empty worker pool")
	return pool
}

// defaultJudgeOrders returns the judge-order rotations assigned to the initial
// workers: the first three workers each prioritize a different judge type so
// no queue can starve, the rest use the default order.
func defaultJudgeOrders(count int) [][]judgetype.JudgeType {
	rotations := [][]judgetype.JudgeType{
		{judgetype.JUDGE, judgetype.JUDGE_PARTIAL, judgetype.REJUDGE},
		{judgetype.JUDGE_PARTIAL, judgetype.REJUDGE, judgetype.JUDGE},
		{judgetype.REJUDGE, judgetype.JUDGE, judgetype.JUDGE_PARTIAL},
	}
	orders := make([][]judgetype.JudgeType, 0, count)
	for i := 0; i < count; i++ {
		if i < len(rotations) {
			orders = append(orders, rotations[i])
			continue
		}
		orders = append(orders, rotations[0])
	}
	return orders
}

// InitJudgeWorkerPool is called as an fx Invoke after startup cleanup has run,
// so orphan runtime resources are removed before workers start picking up work.
func InitJudgeWorkerPool(pool *JudgeWorkerPool, cfg *config.Config) {
	count := cfg.Judge.WorkerCount
	for index, order := range defaultJudgeOrders(count) {
		if err := pool.AddJudgeWorker(order); err != nil {
			pool.logger.Errorf("[JudgeWorkerPool] Failed to create initial worker %d/%d: %v", index+1, count, err)
		}
	}
	pool.logger.Infof("[JudgeWorkerPool] Initialized worker pool with %d workers", len(pool.ListJudgeWorkers()))
}

// ListJudgeWorkers returns a snapshot copy of the pool's workers. A copy, not
// the internal slice: callers (HTTP handlers, orphan services) range over
// the result while AddJudgeWorker/DeleteJudgeWorkerByID can run concurrently
// on other goroutines and would otherwise mutate the same backing array out
// from under an in-progress iteration.
func (p *JudgeWorkerPool) ListJudgeWorkers() []*JudgeWorker {
	p.workersMu.RLock()
	defer p.workersMu.RUnlock()
	workers := make([]*JudgeWorker, len(p.workers))
	copy(workers, p.workers)
	p.logger.Debugf("[JudgeWorkerPool] Listing all workers, count: %d", len(workers))
	return workers
}

func (p *JudgeWorkerPool) AddJudgeWorker(judgeOrder []judgetype.JudgeType) error {
	workerID := uuid.New().String()
	worker := NewJudgeWorker(workerID, p.lock, p.eventbus, p.queueProvider, judgeOrder, p.executor, p.maxJudgeDuration, p.logger)
	p.workersMu.Lock()
	p.workers = append(p.workers, worker)
	p.workersMu.Unlock()
	p.logger.Infof("[JudgeWorkerPool] Added new worker with ID: %s", worker.ID())
	return nil
}

func (p *JudgeWorkerPool) GetJudgeWorkerByID(judgeWorkerID string) *JudgeWorker {
	p.workersMu.RLock()
	defer p.workersMu.RUnlock()
	for _, worker := range p.workers {
		if worker.ID() == judgeWorkerID {
			p.logger.Debugf("[JudgeWorkerPool] Found worker with ID: %s", judgeWorkerID)
			return worker
		}
	}
	p.logger.Warnf("[JudgeWorkerPool] Worker not found with ID: %s", judgeWorkerID)
	return nil
}

func (p *JudgeWorkerPool) DeleteJudgeWorkerByID(judgeWorkerID string) {
	// Remove from the slice under the lock, but run Stop/Close/ReleaseWorker
	// outside it: those can take a while, and holding workersMu across them
	// would block every other List/Get/Add for that whole duration.
	p.workersMu.Lock()
	var target *JudgeWorker
	for i, worker := range p.workers {
		if worker.ID() == judgeWorkerID {
			target = worker
			p.workers = append(p.workers[:i], p.workers[i+1:]...)
			break
		}
	}
	p.workersMu.Unlock()

	if target == nil {
		return
	}

	p.logger.Infof("[JudgeWorkerPool] Deleting worker with ID: %s", judgeWorkerID)
	// markDeletedAndStop, not Stop: a Start() call already holding this
	// worker's handle (from GetJudgeWorkerByID, before this delete) must
	// never be able to revive it after it's been dropped from the pool —
	// see markDeletedAndStop's doc comment.
	target.markDeletedAndStop()
	target.Reader().Close()
	p.provisioner.ReleaseWorker(judgeWorkerID)
	p.logger.Infof("[JudgeWorkerPool] Successfully deleted worker with ID: %s", judgeWorkerID)
}

// WaitForIdle blocks until no worker is running a judgment or the timeout
// elapses. Returns true when all workers are idle. Test seam.
//
// Busyness is detected via IsBusy, which covers the RUNNING status, a
// claimed judgment, and the brief window in between where a worker has
// popped a judgment but not yet claimed it — see JudgeWorker.IsBusy.
func (p *JudgeWorkerPool) WaitForIdle(timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for {
		busy := false
		for _, worker := range p.ListJudgeWorkers() {
			if worker.IsBusy() {
				busy = true
				break
			}
		}
		if !busy {
			return true
		}
		if time.Now().After(deadline) {
			p.logger.Warnf("[JudgeWorkerPool] WaitForIdle timed out after %v", timeout)
			return false
		}
		time.Sleep(50 * time.Millisecond)
	}
}
