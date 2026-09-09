package getproject

import (
	"errors"
	"judge-backend/internal/app/project_domain/project/application/query"
	"judge-backend/internal/app/project_domain/project/application/query/dto"
)

var (
	ErrorInvalidID       = errors.New("invalid id, must be greater than 0")
	ErrorProjectNotFound = errors.New("project not found")
	ErrorInternalServer  = errors.New("internal server error")
)

type Input struct {
	ID int32
}

type Output struct {
	Project *dto.ProjectDetail
}

type Usecase interface {
	Execute(input *Input) (*Output, error)
}

type usecase struct {
	query query.ProjectQuery
}

func NewUsecase(
	query query.ProjectQuery,
) Usecase {
	return &usecase{
		query: query,
	}
}

func (u *usecase) validate(input *Input) error {
	if input.ID < 1 {
		return ErrorInvalidID
	}

	return nil
}

func (u *usecase) Execute(input *Input) (*Output, error) {
	if err := u.validate(input); err != nil {
		return nil, err
	}

	project, err := u.query.GetProjectByID(input.ID)
	if err != nil {
		return nil, ErrorInternalServer
	}

	if project == nil {
		return nil, ErrorProjectNotFound
	}

	return &Output{
		Project: project,
	}, nil
}
