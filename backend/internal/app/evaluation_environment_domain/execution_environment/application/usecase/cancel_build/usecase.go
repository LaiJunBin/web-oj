package cancelbuild

import (
	"errors"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/repository"
	buildservice "judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/service/build_service"
)

var (
	ErrorInvalidInput       = errors.New("tag is required")
	ErrorInternalServer     = errors.New("internal server error")
	ErrorBuildNotFound      = errors.New("build not found")
	ErrorBuildNotInProgress = errors.New("build not in progress")
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
	repo    repository.BuildRepository
	service buildservice.BuildService
}

func NewUsecase(
	repo repository.BuildRepository,
	service buildservice.BuildService,
) Usecase {
	return &usecase{
		repo:    repo,
		service: service,
	}
}

func (u *usecase) validate(input *Input) error {
	if input.Tag == "" {
		return ErrorInvalidInput
	}

	return nil
}

func (u *usecase) Execute(input *Input) (*Output, error) {
	if err := u.validate(input); err != nil {
		return nil, err
	}

	build, err := u.repo.FindLastBuild(input.Tag)
	if err != nil {
		return nil, ErrorInternalServer
	}

	if build == nil {
		return nil, ErrorBuildNotFound
	}

	if !build.IsPending() && !build.IsBuilding() {
		return nil, ErrorBuildNotInProgress
	}

	u.service.Cancel(build)

	return &Output{}, nil
}
