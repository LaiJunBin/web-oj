package provider

import "time"

// WorkerLog is one log line of a live worker.
type WorkerLog struct {
	Text string
	Time time.Time
}

// WorkerHandle is a live judge worker as seen by the worker-management
// subdomain. Implemented by the judge subdomain's worker pool, bridged at the
// composition root, so this subdomain never imports judge infrastructure.
type WorkerHandle interface {
	ID() string
	Status() string
	JudgeOrder() []string
	CreatedAt() time.Time
	LastExecutionAt() *time.Time
	Logs() []WorkerLog
	Start()
	Stop()
}

// WorkerPool manages the live judge workers.
type WorkerPool interface {
	List() []WorkerHandle
	// Get returns nil when no worker has the given ID.
	Get(id string) WorkerHandle
	Add(judgeOrder []string) error
	Delete(id string)
}
