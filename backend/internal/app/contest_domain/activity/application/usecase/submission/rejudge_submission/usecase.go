package rejudgesubmission

import (
	"judge-backend/internal/app/contest_domain/activity/domain/repository"
	"judge-backend/internal/app/contest_domain/activity/domain/service"
	"judge-backend/internal/app/shared/types/httperror"
	"net/http"
)

var (
	ErrorRequiredSubmissionID = httperror.New(http.StatusBadRequest, "Required submission ID")
	ErrorRequiredReason       = httperror.New(http.StatusBadRequest, "Required reason")
	ErrorSubmissionNotFound   = httperror.New(http.StatusNotFound, "Submission not found")
	ErrorInternalServer       = httperror.New(http.StatusInternalServerError, "Internal server error")

	// 400 rather than 409 on purpose: the spec these operations generate from
	// declares 400/401/403/404/500/501 and nothing else, so a 409 would be a
	// response the generated client has no case for.
	ErrorRejudgeAlreadyAwaitingDecision = httperror.New(http.StatusBadRequest, "Submission already has a rejudge awaiting decision")
)

type Input struct {
	SubmissionID string
	Reason       string
}

type Output struct {
}

type Usecase interface {
	Execute(input *Input) (*Output, error)
}

type usecase struct {
	repository        repository.SubmissionRepository
	rejudgeRepository repository.RejudgeRepository
	service           service.SubmissionService
}

func NewUsecase(
	repository repository.SubmissionRepository,
	rejudgeRepository repository.RejudgeRepository,
	service service.SubmissionService,
) Usecase {
	return &usecase{
		repository:        repository,
		rejudgeRepository: rejudgeRepository,
		service:           service,
	}
}

func (u *usecase) validate(input *Input) error {
	if input.SubmissionID == "" {
		return ErrorRequiredSubmissionID
	}

	if input.Reason == "" {
		return ErrorRequiredReason
	}

	return nil
}

func (u *usecase) Execute(input *Input) (*Output, error) {
	if err := u.validate(input); err != nil {
		return nil, err
	}

	submission, err := u.repository.FindByID(input.SubmissionID)
	if err != nil {
		return nil, err
	}

	// Refuse to stack a second rejudge on a submission whose previous one is still
	// undecided. Two outstanding rejudges can be approved in either order, and
	// approving the older one afterwards silently replaces the newer result with a
	// staler score. Cancelling the outstanding rejudge is the way to change course.
	awaiting, err := u.rejudgeRepository.FindAwaitingDecisionBySubmissionIDs([]string{input.SubmissionID})
	if err != nil {
		return nil, err
	}
	if len(awaiting) > 0 {
		return nil, ErrorRejudgeAlreadyAwaitingDecision
	}

	u.service.RequestRejudgeSubmission(submission, input.Reason)

	return &Output{}, nil
}
