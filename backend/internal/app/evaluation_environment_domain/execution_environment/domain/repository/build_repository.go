package repository

import (
	"errors"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/build"
)

var (
	ErrorSaveBuild         = errors.New("error save build")
	ErrorResolveBuildTag   = errors.New("error resolve build tag")
	ErrorResolveImageTag   = errors.New("error resolve image tag")
	ErrorResolveBuildModel = errors.New("error resolve build model")
	// ErrorBuildAlreadyInProgress is returned by SavePendingExclusive when a
	// build for the same tag is still pending or building — enqueueing a
	// second one would overwrite the live build's map entry and orphan it.
	ErrorBuildAlreadyInProgress = errors.New("a build for this tag is already pending or in progress")
)

type BuildRepository interface {
	GetMaxBatch(tag string) (int, error)
	FindByTagAndBatch(tag string, batch int) (*build.Build, error)
	FindLastBuild(tag string) (*build.Build, error)
	FindCurrentBuildByTag(tag string) *build.Build
	// SavePendingExclusive registers a freshly-created pending build, but
	// only if no build for the same tag is currently pending or building —
	// the check and the registration are one atomic step, closing the
	// concurrent-enqueue TOCTOU where two requests both allocated a batch
	// and the second silently overwrote the first's pending entry. Returns
	// ErrorBuildAlreadyInProgress when a live build exists.
	SavePendingExclusive(build *build.Build) error
	// ReleasePending releases a reservation taken by SavePendingExclusive
	// whose follow-up work (project setup, enqueueing) failed, so the next
	// request can reserve the tag again. Identity-guarded: it only removes
	// the entry while it still IS this build, so a late release can never
	// evict a newer reservation.
	ReleasePending(build *build.Build)
	Commit(build *build.Build) error
}
