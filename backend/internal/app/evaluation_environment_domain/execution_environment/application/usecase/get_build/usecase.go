package getbuild

import (
	"errors"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/repository"
	"time"
)

var (
	ErrorTagIsRequired              = errors.New("tag is required")
	ErrorBatchMustBeGreaterThanZero = errors.New("batch must be greater than 0")
	ErrorBuildNotFound              = errors.New("build not found")
	ErrorInternalServer             = errors.New("internal server error")
)

type Input struct {
	Tag   string
	Batch int
}

type Log struct {
	Level   string
	Message string
	Time    time.Time
}

type Output struct {
	Batch  int32
	Status string
	Logs   []Log
}

type Usecase interface {
	Execute(input *Input) (*Output, error)
}

type usecase struct {
	repo repository.BuildRepository
}

func NewUsecase(
	repo repository.BuildRepository,
) Usecase {
	return &usecase{
		repo: repo,
	}
}

func (usecase *usecase) validate(input *Input) error {
	if input.Tag == "" {
		return ErrorTagIsRequired
	}

	if input.Batch <= 0 {
		return ErrorBatchMustBeGreaterThanZero
	}
	return nil
}

func (u *usecase) Execute(input *Input) (*Output, error) {
	if err := u.validate(input); err != nil {
		return nil, err
	}

	build, err := u.repo.FindByTagAndBatch(input.Tag, input.Batch)
	if err != nil {
		return nil, ErrorInternalServer
	}

	if build == nil {
		return nil, ErrorBuildNotFound
	}

	logs := []Log{}
	for _, l := range build.Logs() {
		logs = append(logs, Log{
			Message: l.Message(),
			Time:    l.Time(),
			Level:   l.Level().String(),
		})
	}

	return &Output{
		Batch:  int32(build.Batch().Value()),
		Status: build.Status().String(),
		Logs:   logs,
	}, nil
}
