package worker

import (
	"fmt"
	"time"

	judgeprovider "judge-backend/internal/app/judge_domain/judge/domain/provider"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/judgetype"
)

func parseJudgeOrder(judgeOrder []string) ([]judgetype.JudgeType, error) {
	order := make([]judgetype.JudgeType, 0, len(judgeOrder))
	for _, o := range judgeOrder {
		v, ok := judgetype.JudgeTypeMap[o]
		if !ok {
			return nil, fmt.Errorf("invalid judge order: %s", o)
		}
		order = append(order, v)
	}
	return order, nil
}

// WorkerPoolProvider implements the judge subdomain's OWN published
// contract (domain/provider.WorkerPool) on top of the concrete pool. The
// translation from judge-internal vocabulary (judgetype, internal log
// entries) into the published language happens here, on the judge side, so
// downstream consumers only ever see the published DTOs. The judge
// subdomain knows nothing about who consumes the contract.
type WorkerPoolProvider struct {
	pool *JudgeWorkerPool
}

func NewWorkerPoolProvider(pool *JudgeWorkerPool) judgeprovider.WorkerPool {
	return &WorkerPoolProvider{pool: pool}
}

func (p *WorkerPoolProvider) List() []judgeprovider.WorkerHandle {
	workers := p.pool.ListJudgeWorkers()
	handles := make([]judgeprovider.WorkerHandle, 0, len(workers))
	for _, w := range workers {
		handles = append(handles, &workerHandle{worker: w})
	}
	return handles
}

func (p *WorkerPoolProvider) Get(id string) judgeprovider.WorkerHandle {
	w := p.pool.GetJudgeWorkerByID(id)
	if w == nil {
		return nil
	}
	return &workerHandle{worker: w}
}

func (p *WorkerPoolProvider) Add(judgeOrder []string) error {
	order, err := parseJudgeOrder(judgeOrder)
	if err != nil {
		return err
	}
	return p.pool.AddJudgeWorker(order)
}

func (p *WorkerPoolProvider) Delete(id string) {
	p.pool.DeleteJudgeWorkerByID(id)
}

type workerHandle struct {
	worker *JudgeWorker
}

func (h *workerHandle) ID() string     { return h.worker.ID() }
func (h *workerHandle) Status() string { return h.worker.Status() }

func (h *workerHandle) JudgeOrder() []string {
	order := make([]string, 0, len(h.worker.JudgeOrder()))
	for _, o := range h.worker.JudgeOrder() {
		order = append(order, o.String())
	}
	return order
}

func (h *workerHandle) CreatedAt() time.Time        { return h.worker.CreatedAt() }
func (h *workerHandle) LastExecutionAt() *time.Time { return h.worker.LastExecutionAt() }

func (h *workerHandle) Logs() []judgeprovider.WorkerLog {
	logs := h.worker.Logs()
	result := make([]judgeprovider.WorkerLog, 0, len(logs))
	for _, l := range logs {
		result = append(result, judgeprovider.WorkerLog{Text: l.Text, Time: l.Time})
	}
	return result
}

func (h *workerHandle) Start() { h.worker.Start() }
func (h *workerHandle) Stop()  { h.worker.Stop() }
