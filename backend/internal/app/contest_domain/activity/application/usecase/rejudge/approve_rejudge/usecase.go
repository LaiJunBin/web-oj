package approverejudge

import (
	"judge-backend/internal/app/shared/types/httperror"
	"judge-backend/internal/app/contest_domain/activity/domain/repository"
	"judge-backend/internal/app/contest_domain/activity/domain/service"
	"net/http"
)

var (
	ErrorRejudgeIDIsRequired     = httperror.New(http.StatusBadRequest, "Rejudge ID is required")
	ErrorRejudgeCannotBeApproved = httperror.New(http.StatusBadRequest, "Rejudge cannot be approved")
)

type Input struct {
	RejudgeID string
}

type Output struct {
}

type Usecase interface {
	Execute(input *Input) (*Output, error)
}

type usecase struct {
	repository repository.RejudgeRepository
	service    service.RejudgeService
}

func NewUsecase(
	repository repository.RejudgeRepository,
	service service.RejudgeService,
) Usecase {
	return &usecase{
		repository: repository,
		service:    service,
	}
}

func (u *usecase) validate(input *Input) error {
	if input.RejudgeID == "" {
		return ErrorRejudgeIDIsRequired
	}

	return nil
}

func (u *usecase) Execute(input *Input) (*Output, error) {
	if err := u.validate(input); err != nil {
		return nil, err
	}

	rejudge, err := u.repository.FindByID(input.RejudgeID)
	if err != nil {
		return nil, err
	}

	if !rejudge.CanApprove() {
		return nil, ErrorRejudgeCannotBeApproved
	}

	u.service.RequestApproveRejudge(rejudge)

	return &Output{}, nil
}
