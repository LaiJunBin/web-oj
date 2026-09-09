package build

import (
	"errors"
	"sync"

	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/build/valueobject/batch"
	buildlog "judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/build/valueobject/log"
	buildstatus "judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/build/valueobject/status"
	imagetag "judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/image/valueobject/tag"
)

var (
	ErrInvalidStatusTransition = errors.New("invalid status transition")
)

// Build is internally synchronized, unlike the other entities in this
// domain: BuildRepository's in-memory buildMap hands the SAME instance to
// every caller of the same tag, and those callers run on different
// goroutines (queue workers executing the build, the cancel handler, the
// API's cancel usecase). SetStatus is therefore an atomic compare-and-set —
// the status can only move along CanChangeTo edges no matter how calls
// interleave, and a rejected transition (error) means another writer got
// there first, which callers must treat as "not mine to decide".
type Build struct {
	tag   string
	batch batch.Batch

	mu     sync.Mutex // guards status and logs
	status buildstatus.Status
	logs   []buildlog.Log
}

type BuildConfig struct {
	Tag    imagetag.Tag
	Batch  batch.Batch
	Status buildstatus.Status
}

func New(config *BuildConfig) *Build {
	return &Build{
		tag:    config.Tag.String(),
		batch:  config.Batch,
		status: config.Status,
		logs:   []buildlog.Log{},
	}
}

func (b *Build) AddLog(log buildlog.Log) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.logs = append(b.logs, log)
}

func (b *Build) Tag() string {
	return b.tag
}

// Logs returns a copy: the executing goroutine keeps appending while e.g.
// the cancel handler commits, and handing out the live slice would let the
// reader race the appends.
func (b *Build) Logs() []buildlog.Log {
	b.mu.Lock()
	defer b.mu.Unlock()
	logs := make([]buildlog.Log, len(b.logs))
	copy(logs, b.logs)
	return logs
}

func (b *Build) Batch() batch.Batch {
	return b.batch
}

func (b *Build) Status() buildstatus.Status {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.status
}

// SetStatus is an atomic compare-and-set: the CanChangeTo check and the
// write happen under one lock, so two goroutines can never both take the
// same transition — the loser gets ErrInvalidStatusTransition and must not
// proceed as if it owned the transition.
func (b *Build) SetStatus(status buildstatus.Status) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if !b.status.CanChangeTo(status) {
		return ErrInvalidStatusTransition
	}

	b.status = status
	return nil
}

func (b *Build) IsPending() bool {
	return b.Status().Equals(buildstatus.PENDING)
}

func (b *Build) IsBuilding() bool {
	return b.Status().Equals(buildstatus.BUILDING)
}

func (b *Build) IsCancelled() bool {
	return b.Status().Equals(buildstatus.CANCELLED)
}

func (b *Build) IsSuccess() bool {
	return b.Status().Equals(buildstatus.SUCCESS)
}

func (b *Build) IsError() bool {
	return b.Status().Equals(buildstatus.ERROR)
}
