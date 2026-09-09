// Package provider holds the judge subdomain's published contracts: the
// capabilities it offers to OTHER bounded contexts. Everything here is
// expressed in published language — primitives and plain DTOs only, never
// judge-internal vocabulary (judgetype, entities, infrastructure types) — so
// a downstream context depends on what judge deliberately exposes, not on
// how judge is built.
package provider

import "time"

// WorkerLog is one log line of a live judge worker.
type WorkerLog struct {
	Text string
	Time time.Time
}

// WorkerHandle is a live judge worker as the judge subdomain publishes it.
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

// WorkerPool is the judge subdomain's published worker-management
// capability: list, inspect, add, start/stop and delete the live workers.
// The judge infrastructure implements it; downstream contexts consume it
// through their own anti-corruption adapters and are never imported by the
// judge subdomain.
type WorkerPool interface {
	List() []WorkerHandle
	// Get returns nil when no worker has the given ID.
	Get(id string) WorkerHandle
	Add(judgeOrder []string) error
	Delete(id string)
}
