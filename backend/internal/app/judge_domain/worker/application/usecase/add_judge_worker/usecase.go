package addjudgeworker

import (
	"judge-backend/internal/app/shared/types/httperror"
	"judge-backend/internal/app/judge_domain/worker/domain/entity/judgeworker"
	"judge-backend/internal/app/judge_domain/worker/domain/repository"
	"net/http"
)

var (
	ErrorJudgeOrderIsRequired = httperror.New(http.StatusBadRequest, "Judge Order is required")
)

type Input struct {
	JudgeOrder []string
}

type Output struct {
	ID string
}

type Usecase interface {
	Execute(input *Input) (*Output, error)
}

type usecase struct {
	repo repository.JudgeWorkerRepository
}

func NewUsecase(
	repo repository.JudgeWorkerRepository,
) Usecase {
	return &usecase{
		repo: repo,
	}
}

func (u *usecase) validate(input *Input) error {
	if input.JudgeOrder == nil {
		return ErrorJudgeOrderIsRequired
	}

	if len(input.JudgeOrder) == 0 {
		return ErrorJudgeOrderIsRequired
	}

	return nil
}

func (u *usecase) Execute(input *Input) (*Output, error) {
	if err := u.validate(input); err != nil {
		return nil, err
	}

	judgeWorker, err := judgeworker.New(input.JudgeOrder)
	if err != nil {
		return nil, httperror.New(http.StatusBadRequest, err.Error())
	}

	if err := u.repo.Save(judgeWorker); err != nil {
		return nil, err
	}

	return &Output{
		ID: judgeWorker.ID(),
	}, nil
}
