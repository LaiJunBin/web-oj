package streamsubmissionsbycontestproject

import (
	"judge-backend/internal/app/contest_domain/activity/application/query"
)

type Input struct {
	ContestProjectID int32
}

type Output struct {
	ContestProjectID int32
}

type Usecase interface {
	Execute(input *Input) (*Output, error)
}

type usecase struct {
	contestProjectQuery query.ContestProjectQuery
}

func NewUsecase(contestProjectQuery query.ContestProjectQuery) Usecase {
	return &usecase{contestProjectQuery: contestProjectQuery}
}

func (u *usecase) Execute(input *Input) (*Output, error) {
	if _, err := u.contestProjectQuery.GetContestProjectByID(input.ContestProjectID); err != nil {
		return nil, err
	}

	return &Output{ContestProjectID: input.ContestProjectID}, nil
}
