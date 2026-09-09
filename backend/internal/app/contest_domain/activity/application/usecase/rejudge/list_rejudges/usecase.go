package listrejudges

import (
	"judge-backend/internal/app/contest_domain/activity/application/query"
	"judge-backend/internal/app/contest_domain/activity/application/query/dto"
)

type Input struct {
}

type Output struct {
	Rejudges []*dto.RejudgeInfo
}

type Usecase interface {
	Execute(input *Input) (*Output, error)
}

type usecase struct {
	query query.RejudgeQuery
}

func NewUsecase(
	query query.RejudgeQuery,
) Usecase {
	return &usecase{
		query: query,
	}
}

func (u *usecase) Execute(input *Input) (*Output, error) {
	rejudges, err := u.query.ListRejudges()
	if err != nil {
		return nil, err
	}

	return &Output{
		Rejudges: rejudges,
	}, nil
}
