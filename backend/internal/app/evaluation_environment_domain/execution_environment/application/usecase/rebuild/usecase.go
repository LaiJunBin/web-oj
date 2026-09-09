package rebuild

import (
	"errors"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/repository"
	buildservice "judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/service/build_service"
	"log"
)

var (
	ErrorTagIsRequired      = errors.New("tag is required")
	ErrorImageNotFound      = errors.New("image not found")
	ErrorBuildInProgress    = errors.New("build in progress")
	ErrorInternalServer     = errors.New("internal server error")
	ErrorImageProjectNotSet = errors.New("image project not set")
)

type Input struct {
	Tag string
}

type Output struct {
}

type Usecase interface {
	Execute(input *Input) (*Output, error)
}

type usecase struct {
	imageRepository repository.ImageRepository
	buildRepository repository.BuildRepository
	buildService    buildservice.BuildService
}

func NewUsecase(
	imageRepository repository.ImageRepository,
	buildRepository repository.BuildRepository,
	buildService buildservice.BuildService,
) Usecase {
	return &usecase{
		imageRepository: imageRepository,
		buildRepository: buildRepository,
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

	build, err := u.buildRepository.FindLastBuild(input.Tag)
	if err != nil {
		return nil, ErrorInternalServer
	}

	if build != nil && (build.IsPending() || build.IsBuilding()) {
		return nil, ErrorBuildInProgress
	}

	if err := u.buildService.EnqueueBuildJob(image, &buildservice.BuildOptions{
		NoCache: true,
	}); err != nil {
		// Never swallow an enqueue failure: returning success here while no
		// job exists tells the user a rebuild is running that never will.
		if errors.Is(err, buildservice.ErrorImageProjectNotSet) {
			return nil, ErrorImageProjectNotSet
		}
		if errors.Is(err, repository.ErrorBuildAlreadyInProgress) {
			return nil, ErrorBuildInProgress
		}
		log.Printf("failed to enqueue build image(%s) job: %v", image.Tag(), err)
		return nil, ErrorInternalServer
	}

	return &Output{}, nil
}
