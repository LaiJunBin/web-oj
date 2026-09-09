package updatelifecycle

import (
	"errors"
	"judge-backend/internal/app/project_domain/project/domain/repository"
)

var (
	ErrorInvalidID                   = errors.New("invalid id")
	ErrorInvalidName                 = errors.New("invalid name, must be at least 1 character")
	ErrorInvalidCommand              = errors.New("invalid command, must be at least 1 character")
	ErrorEmptyInput                  = errors.New("empty input")
	ErrorLifecycleNotFound           = errors.New("lifecycle not found")
	ErrorInvalidType                 = errors.New("invalid type, must be one of before_all, before_each, after_each, after_all")
	ErrorInvalidExecutionEnvironment = errors.New("invalid execution environment, must be one of project, test")
	ErrorInternalServer              = errors.New("internal server error")
)

type Input struct {
	ID                   int32
	Type                 *string
	Name                 *string
	Description          *string
	Command              *string
	ExecutionEnvironment *string
}

type Output struct {
}

type Usecase interface {
	Execute(input *Input) (*Output, error)
}

type usecase struct {
	repository repository.LifecycleRepository
}

func NewUsecase(
	repository repository.LifecycleRepository,
) Usecase {
	return &usecase{
		repository: repository,
	}
}

func (u *usecase) validate(input *Input) error {
	if input.ID <= 0 {
		return ErrorInvalidID
	}

	if input.Type == nil && input.Name == nil && input.Description == nil && input.Command == nil && input.ExecutionEnvironment == nil {
		return ErrorEmptyInput
	}

	return nil
}

func (u *usecase) Execute(input *Input) (*Output, error) {
	if err := u.validate(input); err != nil {
		return nil, err
	}

	lc, err := u.repository.FindByID(int(input.ID))
	if err != nil {
		return nil, ErrorInternalServer
	}

	if lc == nil {
		return nil, ErrorLifecycleNotFound
	}

	if input.Type != nil {
		if err := lc.SetType(*input.Type); err != nil {
			return nil, ErrorInvalidType
		}
	}

	if input.ExecutionEnvironment != nil {
		if err := lc.SetExecutionEnvironment(*input.ExecutionEnvironment); err != nil {
			return nil, ErrorInvalidExecutionEnvironment
		}
	}

	if input.Name != nil {
		if err := lc.SetName(*input.Name); err != nil {
			return nil, ErrorInvalidName
		}
	}

	if input.Description != nil {
		lc.SetDescription(*input.Description)
	}

	if input.Command != nil {
		if err := lc.SetCommand(*input.Command); err != nil {
			return nil, ErrorInvalidCommand
		}
	}

	if err := u.repository.Save(lc); err != nil {
		return nil, ErrorInternalServer
	}

	return &Output{}, nil
}
