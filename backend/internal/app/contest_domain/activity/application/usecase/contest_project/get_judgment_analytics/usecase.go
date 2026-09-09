package getjudgmentanalytics

import (
	"judge-backend/internal/app/contest_domain/activity/application/query"
	"judge-backend/internal/app/contest_domain/activity/application/query/dto"
)

type Input struct {
	ContestProjectID int32
}

type Output struct {
	Analytics *dto.JudgmentAnalytics
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
	analytics, err := u.query.GetJudgmentAnalytics(input.ContestProjectID)
	if err != nil {
		return nil, err
	}

	return &Output{
		Analytics: analytics,
	}, nil
}
