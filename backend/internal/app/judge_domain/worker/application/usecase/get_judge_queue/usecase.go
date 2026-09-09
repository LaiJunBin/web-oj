package getjudgequeue

import (
	"judge-backend/internal/app/judge_domain/worker/application/query"
	"judge-backend/internal/app/judge_domain/worker/application/query/dto"
)

type Input struct {
}

type Output struct {
	Status *dto.JudgeQueueStatus
}

type Usecase interface {
	Execute(input *Input) (*Output, error)
}

type usecase struct {
	query query.JudgeQueueQuery
}

func NewUsecase(
	query query.JudgeQueueQuery,
) Usecase {
	return &usecase{
		query: query,
	}
}

func (u *usecase) Execute(input *Input) (*Output, error) {
	status, err := u.query.GetJudgeQueueStatus()
	if err != nil {
		return nil, err
	}

	return &Output{
		Status: status,
	}, nil
}
