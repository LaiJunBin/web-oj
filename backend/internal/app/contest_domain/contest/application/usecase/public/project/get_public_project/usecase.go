package getpublicproject

import (
	"judge-backend/internal/app/contest_domain/contest/application/query"
	"judge-backend/internal/app/contest_domain/contest/application/query/dto"
	sharedtypes "judge-backend/internal/app/shared/types"
	"judge-backend/internal/app/shared/types/httperror"
	"net/http"
)

var (
	ErrorProjectNotFound = httperror.New(http.StatusNotFound, "Project not found")
)

type Input struct {
	ContestProjectID int32
}

type Output struct {
	Project dto.ProjectDetail
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
	u.logger.Infof("Starting to get public project", "projectId", input.ContestProjectID)

	project, err := u.query.GetPublicProjectByID(input.ContestProjectID)
	if err != nil {
		u.logger.Errorf("Failed to get public project", "projectId", input.ContestProjectID, "error", err)
		return nil, err
	}

	u.logger.Infof("Successfully got public project", "projectId", input.ContestProjectID)
	return &Output{
		Project: project,
	}, nil
}
