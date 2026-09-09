package buildservice

import (
	"errors"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/build"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/image"
)

var (
	ErrorImageProjectNotSet                = errors.New("image project not set")
	ErrorBuildNotFound                     = errors.New("build not found")
	ErrorBuildStatusCanNotChangeToBuilding = errors.New("build status can't change to building")
)

type BuildOptions struct {
	NoCache bool
}

type BuildService interface {
	// EnqueueBuildJob reserves the tag's exclusive build slot and enqueues
	// in one call — for flows with no project mutation between the two
	// (e.g. rebuild). Setup-style flows must use the two-phase methods
	// below instead, so their file writes happen under the reservation.
	EnqueueBuildJob(image *image.Image, options *BuildOptions) error
	// ReserveBuild claims the tag's exclusive pending slot (atomically —
	// concurrent reservations for one tag admit exactly one winner) and
	// returns the reserved build. The caller now owns the slot: it must
	// either enqueue the reserved build or release it.
	ReserveBuild(image *image.Image) (*build.Build, error)
	// EnqueueReservedBuild enqueues a previously reserved build, after the
	// caller finished preparing the project files. On error the caller must
	// release the reservation.
	EnqueueReservedBuild(build *build.Build, image *image.Image, options *BuildOptions) error
	// ReleaseReservedBuild frees a reservation whose setup or enqueue
	// failed (identity-guarded — a late release cannot evict a newer
	// reservation), so the next request can reserve the tag again.
	ReleaseReservedBuild(build *build.Build)
	Cancel(build *build.Build)
}
