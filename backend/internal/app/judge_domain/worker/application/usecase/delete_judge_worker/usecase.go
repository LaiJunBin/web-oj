package deletejudgeworker

import (
	"judge-backend/internal/app/shared/types/httperror"
	"judge-backend/internal/app/judge_domain/worker/domain/repository"
	"net/http"
)

var (
	ErrorJudgeWorkerIDIsRequired = httperror.New(http.StatusBadRequest, "JudgeWorker ID is required")
)

type Input struct {
	JudgeWorkerID string
}

type Output struct {
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
	if input.JudgeWorkerID == "" {
		return ErrorJudgeWorkerIDIsRequired
	}

	return nil
}

func (u *usecase) Execute(input *Input) (*Output, error) {
	if err := u.validate(input); err != nil {
		return nil, err
	}

	judgeWorker, err := u.repo.FindByID(input.JudgeWorkerID)
	if err != nil {
		return nil, httperror.New(http.StatusBadRequest, err.Error())
	}

	if err := u.repo.Delete(judgeWorker); err != nil {
		return nil, err
	}

	return &Output{}, nil
}
