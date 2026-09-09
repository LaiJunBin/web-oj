package streamuserprojectclarification

import (
	"judge-backend/internal/app/contest_domain/contest/application/query"
	"judge-backend/internal/app/contest_domain/contest/domain/repository"
	sharedtypes "judge-backend/internal/app/shared/types"
	"judge-backend/internal/app/shared/types/httperror"
	"net/http"
)

var (
	ErrorInvalidUserID           = httperror.New(http.StatusBadRequest, "invalid user ID")
	ErrorInvalidContestProjectID = httperror.New(http.StatusBadRequest, "invalid contest project ID")
	ErrorContestProjectNotFound  = httperror.New(http.StatusNotFound, "contest project not found")
)

type Input struct {
	UserID           int32
	ContestProjectID int32
}

type Output struct {
	UserID           int32
	ContestProjectID int32
}

type Usecase interface {
	Execute(input *Input) (*Output, error)
}

type usecase struct {
	repository          repository.ProjectRepository
	contestProjectQuery query.ContestProjectQuery
	logger              sharedtypes.Logger
}

func NewUsecase(
	repository repository.ProjectRepository,
	contestProjectQuery query.ContestProjectQuery,
	logger sharedtypes.Logger,
) Usecase {
	return &usecase{
		repository:          repository,
		contestProjectQuery: contestProjectQuery,
		logger:              logger,
	}
}

func (u *usecase) validate(input *Input) error {
	if input.UserID <= 0 {
		return ErrorInvalidUserID
	}

	if input.ContestProjectID <= 0 {
		return ErrorInvalidContestProjectID
	}

	return nil
}

func (u *usecase) Execute(input *Input) (*Output, error) {
	u.logger.Infof("Starting to stream user project clarification", "userId", input.UserID, "projectId", input.ContestProjectID)

	if err := u.validate(input); err != nil {
		u.logger.Errorf("Failed to validate input", "userId", input.UserID, "projectId", input.ContestProjectID, "error", err)
		return nil, err
	}

	if err := u.contestProjectQuery.CheckUserCanAccessProject(input.ContestProjectID, input.UserID); err != nil {
		u.logger.Errorf("User cannot access project clarification stream", "userId", input.UserID, "projectId", input.ContestProjectID, "error", err)
		return nil, err
	}

	project, err := u.repository.FindByID(input.ContestProjectID)
	if err != nil {
		u.logger.Errorf("Failed to find project", "userId", input.UserID, "projectId", input.ContestProjectID, "error", err)
		return nil, err
	}

	if project == nil {
		u.logger.Errorf("Project not found", "userId", input.UserID, "projectId", input.ContestProjectID)
		return nil, ErrorContestProjectNotFound
	}

	u.logger.Infof("Successfully started streaming user project clarification", "userId", input.UserID, "projectId", input.ContestProjectID)
	return &Output{
		UserID:           input.UserID,
		ContestProjectID: input.ContestProjectID,
	}, nil
}
