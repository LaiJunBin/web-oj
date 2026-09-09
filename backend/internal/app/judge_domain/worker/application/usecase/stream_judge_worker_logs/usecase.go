package streamjudgeworkerlogs

import (
	"judge-backend/internal/app/shared/types/httperror"
	"judge-backend/internal/app/judge_domain/worker/application/query"
	"judge-backend/internal/app/judge_domain/worker/application/query/dto"
	"net/http"
)

var (
	ErrorJudgeWorkerIDIsRequired = httperror.New(http.StatusBadRequest, "JudgeWorker ID is required")
)

type Input struct {
	JudgeWorkerID string
}

type Output struct {
	judgeWorker *dto.JudgeWorker
}

func (o *Output) ID() string {
	return o.judgeWorker.ID
}

func (o *Output) InitLog() []*dto.Log {
	return o.judgeWorker.Logs
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

func (u *usecase) validate(input *Input) error {
	if input.JudgeWorkerID == "" {
		return ErrorJudgeWorkerIDIsRequired
	}

	return nil
}

func (u *usecase) Execute(input *Input) (*Output, error) {
	if err := u.validate(input); err != nil {
		return nil, err
	}

	judgeWorker, err := u.query.GetJudgeWorkerByID(input.JudgeWorkerID)
	if err != nil {
		return nil, err
	}

	return &Output{
		judgeWorker: judgeWorker,
	}, nil
}
