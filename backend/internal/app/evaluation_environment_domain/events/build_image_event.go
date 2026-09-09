package event

import (
	"time"
)

type BuildImageOptions struct {
	NoCache bool
}

type BuildImageEvent struct {
	tag       string
	batch     int
	path      string
	options   *BuildImageOptions
	createdAt time.Time
}

type BuildImageEventConfig struct {
	Tag string
	// Batch pins the event to ONE build: tag alone is not an identity —
	// by the time a queued event is consumed, the "last" build of its tag
	// can be a different, newer build entirely.
	Batch   int
	Path    string
	Options *BuildImageOptions
}

func (e *BuildImageEvent) Topic() string {
	return "build_image"
}

func NewBuildImageEvent(config *BuildImageEventConfig) *BuildImageEvent {
	return &BuildImageEvent{
		tag:       config.Tag,
		batch:     config.Batch,
		path:      config.Path,
		options:   config.Options,
		createdAt: time.Now(),
	}
}

func (e *BuildImageEvent) Tag() string {
	return e.tag
}

func (e *BuildImageEvent) Batch() int {
	return e.batch
}

func (e *BuildImageEvent) Path() string {
	return e.path
}

func (e *BuildImageEvent) Options() *BuildImageOptions {
	return e.options
}

func (e *BuildImageEvent) CreatedAt() time.Time {
	return e.createdAt
}
