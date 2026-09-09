package deletelifecycle

import (
	"errors"
	"judge-backend/internal/app/project_domain/project/domain/repository"
)

var (
	ErrorInvalidID         = errors.New("invalid id")
	ErrorLifecycleNotFound = errors.New("lifecycle not found")
	ErrorInternalServer    = errors.New("internal server error")
)

type Input struct {
	ID int32
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

	if err := u.repository.Delete(lc); err != nil {
		return nil, ErrorInternalServer
	}

	return &Output{}, nil
}
