package listuserprojectsubmissions

import (
	"judge-backend/internal/app/contest_domain/contest/application/query"
	"judge-backend/internal/app/contest_domain/contest/application/query/dto"
	sharedtypes "judge-backend/internal/app/shared/types"
)

type Input struct {
	UserID           int32
	ContestProjectID int32
}

type Output struct {
	Submissions []*dto.Submission
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
	u.logger.Infof("Starting to list user project submissions", "userId", input.UserID, "projectId", input.ContestProjectID)

	submissions, err := u.query.ListUserProjectSubmissionsByProjectID(input.UserID, input.ContestProjectID)
	if err != nil {
		u.logger.Errorf("Failed to list user project submissions", "userId", input.UserID, "projectId", input.ContestProjectID, "error", err)
		return nil, err
	}

	u.logger.Infof("Successfully listed user project submissions", "userId", input.UserID, "projectId", input.ContestProjectID)
	return &Output{
		Submissions: submissions,
	}, nil
}
