package listproject

import (
	"errors"
	"judge-backend/internal/app/project_domain/project/application/query"
	"judge-backend/internal/app/project_domain/project/application/query/dto"
)

var (
	ErrorInternalServer = errors.New("internal server error")
)

type Input struct {
}

type Output struct {
	Projects []dto.Project
}

type Usecase interface {
	Execute(input *Input) (*Output, error)
}

type usecase struct {
	query query.ProjectQuery
}

func NewUsecase(query query.ProjectQuery) Usecase {
	return &usecase{
		query: query,
	}
}

func (u *usecase) Execute(input *Input) (*Output, error) {
	projects, err := u.query.ListProject()
	if err != nil {
		return nil, ErrorInternalServer
	}

	return &Output{
		Projects: projects,
	}, nil
}
