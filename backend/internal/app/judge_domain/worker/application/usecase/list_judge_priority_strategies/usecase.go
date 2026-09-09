package listjudgeprioritystrategies

import (
	"judge-backend/internal/app/judge_domain/worker/application/query"
)

type Input struct {
}

type Output struct {
	JudgePriorityStrategies []string
}

type Usecase interface {
	Execute(input *Input) (*Output, error)
}

type usecase struct {
	query query.JudgeWorkerQuery
}

func NewUsecase(
	query query.JudgeWorkerQuery,
) Usecase {
	return &usecase{
		query: query,
	}
}

func (u *usecase) Execute(input *Input) (*Output, error) {
	judgePriorityStrategies, err := u.query.ListJudgePriorityStrategies()
	if err != nil {
		return nil, err
	}

	return &Output{
		JudgePriorityStrategies: judgePriorityStrategies,
	}, nil
}
