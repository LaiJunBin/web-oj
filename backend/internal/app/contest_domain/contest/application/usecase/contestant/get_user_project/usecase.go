package getuserproject

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
	Project dto.UserProjectDetail
}

type Usecase interface {
	Execute(input *Input) (*Output, error)
}

type usecase struct {
	query  query.ProjectQuery
	logger sharedtypes.Logger
}

func NewUsecase(
	query query.ProjectQuery,
	logger sharedtypes.Logger,
) Usecase {
	return &usecase{
		query:  query,
		logger: logger,
	}
}

func (u *usecase) Execute(input *Input) (*Output, error) {
	u.logger.Infof("Starting to get user project", "userId", input.UserID, "projectId", input.ContestProjectID)

	project, err := u.query.GetUserProjectByID(input.UserID, input.ContestProjectID)
	if err != nil {
		u.logger.Errorf("Failed to get user project from database", "userId", input.UserID, "projectId", input.ContestProjectID, "error", err)
		return nil, err
	}

	u.logger.Infof("Successfully got user project", "userId", input.UserID, "projectId", input.ContestProjectID)
	return &Output{
		Project: project,
	}, nil
}
