package updateimage

import (
	"bytes"
	"errors"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/repository"
	buildservice "judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/service/build_service"
	imageservice "judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/service/image_service"
	"log"
)

var (
	ErrorInvalidInput           = errors.New("invalid input")
	ErrorInvalidMountPath       = errors.New("invalid mount path")
	ErrorTagIsRequired          = errors.New("tag is required")
	ErrorImageNotFound          = errors.New("image not found")
	ErrorInvalidTag             = errors.New("invalid tag")
	ErrorCurrentBuildInProgress = errors.New("current build in progress, please cancel the build first")
	ErrorFailedToSetup          = errors.New("failed to setup project")
	ErrorInternalServer         = errors.New("internal server error")
)

type Input struct {
	Tag            string
	MountPath      *string
	Description    *string
	ProjectZipFile *bytes.Buffer
	NoCache        bool
}

type Output struct{}

type Usecase interface {
	Execute(input *Input) (*Output, error)
}

type usecase struct {
	imageRepository repository.ImageRepository
	buildRepository repository.BuildRepository
	imageService    imageservice.ImageService
	buildService    buildservice.BuildService
}

func NewUsecase(
	imageRepository repository.ImageRepository,
	buildRepository repository.BuildRepository,
	imageService imageservice.ImageService,
	buildService buildservice.BuildService,
) Usecase {
	return &usecase{
		imageRepository: imageRepository,
		buildRepository: buildRepository,
		imageService:    imageService,
		buildService:    buildService,
	}
}

func (u *usecase) validate(input *Input) error {
	if input.Tag == "" {
		return ErrorTagIsRequired
	}

	return nil
}

func (u *usecase) Execute(input *Input) (*Output, error) {
	if err := u.validate(input); err != nil {
		return nil, err
	}

	image, err := u.imageRepository.FindByTag(input.Tag)
	if err != nil {
		return nil, ErrorInternalServer
	}

	if image == nil {
		return nil, ErrorImageNotFound
	}

	if input.Description != nil {
		image.SetDescription(*input.Description)
	}

	if input.MountPath != nil {
		if err := image.SetMountPath(*input.MountPath); err != nil {
			return nil, ErrorInvalidMountPath
		}
	}

	if input.ProjectZipFile != nil {
		// DB-visible precheck first: the reservation map only knows about
		// builds live in THIS process, while a build row can read
		// Pending/Building from seeded or crash-orphaned state the map never
		// saw. FindLastBuild consults both (map, then DB), so this rejects
		// those; the reservation below then closes the in-process TOCTOU
		// this check alone cannot.
		lastBuild, err := u.buildRepository.FindLastBuild(input.Tag)
		if err != nil {
			return nil, ErrorInternalServer
		}
		if lastBuild != nil && (lastBuild.IsPending() || lastBuild.IsBuilding()) {
			return nil, ErrorCurrentBuildInProgress
		}

		// Two-phase: reserve the tag's exclusive build slot FIRST, and only
		// then touch the project files. SetupProject overwrites the files a
		// concurrently-running build would read — with the old order (check,
		// setup, enqueue) two racing updates could interleave their writes,
		// and the winning build could build the loser's content. The
		// reservation is the same one every enqueue path goes through, so
		// this excludes rebuild/create/update uniformly, not just other
		// updates.
		reserved, err := u.buildService.ReserveBuild(image)
		if err != nil {
			if errors.Is(err, repository.ErrorBuildAlreadyInProgress) {
				return nil, ErrorCurrentBuildInProgress
			}
			return nil, ErrorInternalServer
		}

		if err := u.imageService.SetupProject(image, input.ProjectZipFile); err != nil {
			// Release, or the tag stays reserved forever and no later
			// update/rebuild could ever enqueue again.
			u.buildService.ReleaseReservedBuild(reserved)
			return nil, ErrorFailedToSetup
		}

		if err := u.buildService.EnqueueReservedBuild(reserved, image, &buildservice.BuildOptions{
			NoCache: input.NoCache,
		}); err != nil {
			log.Printf("failed to enqueue build image(%s) job: %v", image.Tag(), err)
			u.buildService.ReleaseReservedBuild(reserved)
			return nil, ErrorInternalServer
		}
	}

	if err := u.imageRepository.Save(image); err != nil {
		return nil, ErrorInternalServer
	}

	return &Output{}, nil
}
