package getlifecycle

import (
	"errors"
	"judge-backend/internal/app/project_domain/project/application/query"
	"judge-backend/internal/app/project_domain/project/application/query/dto"
)

var (
	ErrorInvalidID         = errors.New("invalid id, id must be greater than 0")
	ErrorLifecycleNotFound = errors.New("lifecycle not found")
	ErrorInternalServer    = errors.New("internal server error")
)

type Input struct {
	ID int32
}

type Output struct {
	Lifecycle *dto.LifecycleCommand
}

type Usecase interface {
	Execute(input *Input) (*Output, error)
}

type usecase struct {
	query query.LifecycleQuery
}

func NewUsecase(
	query query.LifecycleQuery,
) Usecase {
	return &usecase{
		query: query,
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

	lifecycle, err := u.query.GetLifecycle(input.ID)
	if err != nil {
		return nil, ErrorInternalServer
	}

	if lifecycle == nil {
		return nil, ErrorLifecycleNotFound
	}

	return &Output{
		Lifecycle: lifecycle,
	}, nil
}
