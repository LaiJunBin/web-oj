package streambuildlog

import (
	"errors"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/build"
	buildlog "judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/build/valueobject/log"
	buildstatus "judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/build/valueobject/status"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/repository"
)

var (
	ErrorTagIsRequired              = errors.New("tag is required")
	ErrorBatchMustBeGreaterThanZero = errors.New("batch must be greater than zero")
	ErrorImageNotFound              = errors.New("image not found")
	ErrorInternalServer             = errors.New("internal server error")
	ErrorBuildNotFound              = errors.New("build not found")
	ErrorBuildNotInProgress         = errors.New("build not in progress")
	ErrorBuildNotStarted            = errors.New("build not started")
)

type Usecase interface {
	Execute(input *Input) (*Output, error)
}

type Input struct {
	Tag   string
	Batch int
}

type Output struct {
	build *build.Build
}

func (o *Output) Logs() []buildlog.Log {
	return o.build.Logs()
}

func (o *Output) Status() buildstatus.Status {
	return o.build.Status()
}

type usecase struct {
	imageRepo repository.ImageRepository
	buildRepo repository.BuildRepository
}

func NewUsecase(
	imageRepo repository.ImageRepository,
	buildRepo repository.BuildRepository,
) Usecase {
	return &usecase{
		imageRepo: imageRepo,
		buildRepo: buildRepo,
	}
}

func (u *usecase) Validate(input *Input) error {
	if input.Tag == "" {
		return ErrorTagIsRequired
	}

	if input.Batch <= 0 {
		return ErrorBatchMustBeGreaterThanZero
	}

	return nil
}

func (u *usecase) Execute(input *Input) (*Output, error) {
	if err := u.Validate(input); err != nil {
		return nil, err
	}

	build, err := u.buildRepo.FindByTagAndBatch(input.Tag, input.Batch)
	if err != nil {
		return nil, ErrorInternalServer
	}

	if build == nil {
		return nil, ErrorBuildNotFound
	}

	return &Output{
		build: build,
	}, nil
}
