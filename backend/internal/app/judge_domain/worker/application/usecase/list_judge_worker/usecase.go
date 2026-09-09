package listjudgeworker

import (
	"judge-backend/internal/app/judge_domain/worker/application/query"
	"judge-backend/internal/app/judge_domain/worker/application/query/dto"
)

type Input struct {
}

type Output struct {
	JudgeWorkers []*dto.JudgeWorkerInfo
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
	judgeWorkers, err := u.query.ListJudgeWorker()
	if err != nil {
		return nil, err
	}

	return &Output{
		JudgeWorkers: judgeWorkers,
	}, nil
}
