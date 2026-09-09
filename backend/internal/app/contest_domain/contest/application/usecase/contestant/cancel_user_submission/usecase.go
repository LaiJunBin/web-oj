package cancelusersubmission

import (
	"judge-backend/internal/app/contest_domain/contest/domain/repository"
	"judge-backend/internal/app/contest_domain/contest/domain/service"
	sharedtypes "judge-backend/internal/app/shared/types"
	"judge-backend/internal/app/shared/types/httperror"
	"net/http"
)

var (
	ErrorInvalidUserID           = httperror.New(http.StatusBadRequest, "invalid user id")
	ErrorInvalidSubmissionID     = httperror.New(http.StatusBadRequest, "invalid submission id")
	ErrorForbidden               = httperror.New(http.StatusForbidden, "forbidden")
	ErrorInvalidSubmissionStatus = httperror.New(http.StatusBadRequest, "invalid submission status")
)

type Input struct {
	UserID       int32
	SubmissionID string
}

type Output struct {
}

type Usecase interface {
	Execute(input *Input) (*Output, error)
}

type usecase struct {
	repository repository.SubmissionRepository
	service    service.SubmissionService
	logger     sharedtypes.Logger
}

func NewUsecase(
	repository repository.SubmissionRepository,
	service service.SubmissionService,
	logger sharedtypes.Logger,
) Usecase {
	return &usecase{
		repository: repository,
		service:    service,
		logger:     logger,
	}
}

func (u *usecase) validate(input *Input) error {
	if input.UserID <= 0 {
		return ErrorInvalidUserID
	}

	if input.SubmissionID == "" {
		return ErrorInvalidSubmissionID
	}

	return nil
}

func (u *usecase) Execute(input *Input) (*Output, error) {
	u.logger.Infof("Starting to cancel user submission", "userId", input.UserID, "submissionId", input.SubmissionID)

	if err := u.validate(input); err != nil {
		u.logger.Errorf("Failed to validate input", "userId", input.UserID, "submissionId", input.SubmissionID, "error", err)
		return nil, err
	}

	submission, err := u.repository.FindByID(input.SubmissionID)
	if err != nil {
		u.logger.Errorf("Failed to find submission by ID", "userId", input.UserID, "submissionId", input.SubmissionID, "error", err)
		return nil, err
	}

	if submission.UserID() != input.UserID {
		u.logger.Errorf("User is not authorized to cancel this submission", "userId", input.UserID, "submissionId", input.SubmissionID)
		return nil, ErrorForbidden
	}

	if !submission.CanCancel() {
		u.logger.Errorf("Submission cannot be cancelled", "userId", input.UserID, "submissionId", input.SubmissionID)
		return nil, ErrorInvalidSubmissionStatus
	}

	u.service.RequestCancelSubmission(submission)
	u.logger.Infof("Successfully requested to cancel user submission", "userId", input.UserID, "submissionId", input.SubmissionID)

	return nil, nil
}
