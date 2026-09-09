package canceljudge

import (
	"judge-backend/internal/app/shared/types/httperror"
	"judge-backend/internal/app/contest_domain/activity/domain/repository"
	"judge-backend/internal/app/contest_domain/activity/domain/service"
	"net/http"
)

var (
	ErrorJudgmentIDIsRequired      = httperror.New(http.StatusBadRequest, "Judgment ID is required")
	ErrorJudgmentCannotBeCancelled = httperror.New(http.StatusBadRequest, "Judgment cannot be cancelled")
)

type Input struct {
	JudgmentID string
	Reason     string
}

type Output struct {
}

type Usecase interface {
	Execute(input *Input) (*Output, error)
}

type usecase struct {
	repository repository.JudgmentRepository
	service    service.JudgmentService
}

func NewUsecase(
	repository repository.JudgmentRepository,
	service service.JudgmentService,
) Usecase {
	return &usecase{
		repository: repository,
		service:    service,
	}
}

func (u *usecase) validate(input *Input) error {
	if input.JudgmentID == "" {
		return ErrorJudgmentIDIsRequired
	}

	if input.Reason == "" {
		// Contestants see the cancellation reason on their submission, so the
		// default an admin leaves in place has to read as a message to them.
		input.Reason = "評測已由工作人員取消"
	}

	return nil
}

func (u *usecase) Execute(input *Input) (*Output, error) {
	if err := u.validate(input); err != nil {
		return nil, err
	}

	judgment, err := u.repository.FindByID(input.JudgmentID)
	if err != nil {
		return nil, err
	}

	if !judgment.CanCancel() {
		return nil, ErrorJudgmentCannotBeCancelled
	}

	u.service.RequestCancelJudgeByID(judgment.ID(), input.Reason)

	return &Output{}, nil
}
