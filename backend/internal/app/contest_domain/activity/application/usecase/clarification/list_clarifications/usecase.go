package listclarifications

import (
	"judge-backend/internal/app/contest_domain/activity/application/query"
	"judge-backend/internal/app/contest_domain/activity/application/query/dto"
)

type Input struct {
	ContestProjectID int32
}

type Output struct {
	Clarifications []*dto.Clarification
}

type Usecase interface {
	Execute(input *Input) (*Output, error)
}

type usecase struct {
	query query.ClarificationQuery
}

func NewUsecase(
	query query.ClarificationQuery,
) Usecase {
	return &usecase{
		query: query,
	}
}

func (u *usecase) Execute(input *Input) (*Output, error) {
	clarifications, err := u.query.ListClarifications(input.ContestProjectID)
	if err != nil {
		return nil, err
	}

	return &Output{
		Clarifications: clarifications,
	}, nil
}
