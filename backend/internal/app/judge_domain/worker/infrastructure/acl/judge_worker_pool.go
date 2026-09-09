// Package acl is the worker subdomain's anti-corruption layer: the ONE
// place in this subdomain that knows about the judge subdomain. It consumes
// judge's published contract (judge/domain/provider) and translates it into
// this subdomain's own port (worker/domain/provider), so every inner layer
// of the worker subdomain depends exclusively on its own vocabulary. Today
// the two contracts are near-identical and the adapter is a thin
// delegation; if they ever diverge — or the judge subdomain moves behind a
// network boundary — this package absorbs the difference and the rest of
// the subdomain does not change.
package acl

import (
	"time"

	judgeprovider "judge-backend/internal/app/judge_domain/judge/domain/provider"
	workerprovider "judge-backend/internal/app/judge_domain/worker/domain/provider"
)

type judgeWorkerPool struct {
	pool judgeprovider.WorkerPool
}

// NewJudgeWorkerPool adapts the judge subdomain's published worker-pool
// contract into this subdomain's WorkerPool port.
func NewJudgeWorkerPool(pool judgeprovider.WorkerPool) workerprovider.WorkerPool {
	return &judgeWorkerPool{pool: pool}
}

func (a *judgeWorkerPool) List() []workerprovider.WorkerHandle {
	upstream := a.pool.List()
	handles := make([]workerprovider.WorkerHandle, 0, len(upstream))
	for _, h := range upstream {
		handles = append(handles, &judgeWorkerHandle{handle: h})
	}
	return handles
}

func (a *judgeWorkerPool) Get(id string) workerprovider.WorkerHandle {
	h := a.pool.Get(id)
	if h == nil {
		return nil
	}
	return &judgeWorkerHandle{handle: h}
}

func (a *judgeWorkerPool) Add(judgeOrder []string) error {
	return a.pool.Add(judgeOrder)
}

func (a *judgeWorkerPool) Delete(id string) {
	a.pool.Delete(id)
}

type judgeWorkerHandle struct {
	handle judgeprovider.WorkerHandle
}

func (h *judgeWorkerHandle) ID() string                  { return h.handle.ID() }
func (h *judgeWorkerHandle) Status() string              { return h.handle.Status() }
func (h *judgeWorkerHandle) JudgeOrder() []string        { return h.handle.JudgeOrder() }
func (h *judgeWorkerHandle) CreatedAt() time.Time        { return h.handle.CreatedAt() }
func (h *judgeWorkerHandle) LastExecutionAt() *time.Time { return h.handle.LastExecutionAt() }

func (h *judgeWorkerHandle) Logs() []workerprovider.WorkerLog {
	upstream := h.handle.Logs()
	logs := make([]workerprovider.WorkerLog, 0, len(upstream))
	for _, l := range upstream {
		logs = append(logs, workerprovider.WorkerLog{Text: l.Text, Time: l.Time})
	}
	return logs
}

func (h *judgeWorkerHandle) Start() { h.handle.Start() }
func (h *judgeWorkerHandle) Stop()  { h.handle.Stop() }
