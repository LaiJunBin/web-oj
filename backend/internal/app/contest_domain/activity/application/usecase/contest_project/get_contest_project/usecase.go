package getcontestproject

import (
	"judge-backend/internal/app/contest_domain/activity/application/query"
	"judge-backend/internal/app/contest_domain/activity/application/query/dto"
)

type Input struct {
	ID int32
}

type Output struct {
	Project *dto.ContestProject
}

type Usecase interface {
	Execute(input *Input) (*Output, error)
}

type usecase struct {
	query query.ContestProjectQuery
}

func NewUsecase(
	query query.ContestProjectQuery,
) Usecase {
	return &usecase{
		query: query,
	}
}

func (u *usecase) Execute(input *Input) (*Output, error) {
	project, err := u.query.GetContestProjectByID(input.ID)
	if err != nil {
		return nil, err
	}

	return &Output{
		Project: project,
	}, nil
}
