package getusersubmission

import (
	"judge-backend/internal/app/contest_domain/contest/application/query"
	"judge-backend/internal/app/contest_domain/contest/application/query/dto"
	sharedtypes "judge-backend/internal/app/shared/types"
)

type Input struct {
	UserID       int32
	SubmissionID string
}

type Output struct {
	Submission *dto.SubmissionDetail
}

type Usecase interface {
	Execute(input *Input) (*Output, error)
}

type usecase struct {
	query  query.SubmissionQuery
	logger sharedtypes.Logger
}

func NewUsecase(
	query query.SubmissionQuery,
	logger sharedtypes.Logger,
) Usecase {
	return &usecase{
		query:  query,
		logger: logger,
	}
}

func (u *usecase) Execute(input *Input) (*Output, error) {
	u.logger.Infof("Starting to get user submission", "userId", input.UserID, "submissionId", input.SubmissionID)

	submission, err := u.query.GetUserSubmissionByID(input.UserID, input.SubmissionID)
	if err != nil {
		u.logger.Errorf("Failed to get user submission", "userId", input.UserID, "submissionId", input.SubmissionID, "error", err)
		return nil, err
	}

	u.logger.Infof("Successfully got user submission", "userId", input.UserID, "submissionId", input.SubmissionID)
	return &Output{
		Submission: submission,
	}, nil
}
