package getuserscores

import (
	"judge-backend/internal/app/contest_domain/activity/application/query"
	"judge-backend/internal/app/contest_domain/activity/application/query/dto"
)

type Input struct {
	ID int32
}

type Output struct {
	UserScores []*dto.UserScore
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
	userScores, err := u.query.GetUserScores(input.ID)
	if err != nil {
		return nil, err
	}

	return &Output{
		UserScores: userScores,
	}, nil
}
