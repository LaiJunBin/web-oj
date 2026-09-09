package deleteproject

import (
	"errors"
	"judge-backend/internal/app/project_domain/project/domain/repository"
)

var (
	ErrorInvalidID       = errors.New("invalid id, id must be greater than 0")
	ErrorProjectNotFound = errors.New("project not found")
	ErrorInternalServer  = errors.New("internal server error")
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
	repository repository.ProjectRepository
}

func NewUsecase(
	repository repository.ProjectRepository,
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

	project, err := u.repository.FindByID(int(input.ID))
	if err != nil {
		return nil, ErrorInternalServer
	}

	if project == nil {
		return nil, ErrorProjectNotFound
	}

	if err := u.repository.Delete(project); err != nil {
		return nil, ErrorInternalServer
	}

	return &Output{}, nil
}
