package cancelrejudge

import (
	"judge-backend/internal/app/shared/types/httperror"
	"judge-backend/internal/app/contest_domain/activity/domain/repository"
	"judge-backend/internal/app/contest_domain/activity/domain/service"
	"net/http"
)

var (
	ErrorRejudgeIDIsRequired       = httperror.New(http.StatusBadRequest, "Rejudge ID is required")
	ErrorJudgmentCannotBeCancelled = httperror.New(http.StatusBadRequest, "Judgment cannot be cancelled")
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

	// An already-CANCELED rejudge accepts the request again as an idempotent
	// retry: a previous cancellation can have flipped the parent and then
	// failed partway through its members, leaving some still queued or
	// running. Republishing the cancel event (each request carries a fresh
	// event ID, so the claim-based dedup never blocks it) lets the
	// judge-domain handler reconcile those stragglers — without this
	// carve-out that recovery path is unreachable. Every other
	// non-cancellable status stays rejected.
	if !rejudge.CanCancel() && !rejudge.Status().IsCanceled() {
		return nil, ErrorJudgmentCannotBeCancelled
	}

	u.service.RequestCancelRejudge(rejudge)

	return &Output{}, nil
}
