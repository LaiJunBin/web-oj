package streamuserprojecttime

import (
	"judge-backend/internal/app/contest_domain/contest/application/query"
	"judge-backend/internal/app/contest_domain/contest/domain/entity/project"
	"judge-backend/internal/app/contest_domain/contest/domain/repository"
	sharedtypes "judge-backend/internal/app/shared/types"
	"judge-backend/internal/app/shared/types/httperror"
	"net/http"
	"time"
)

var (
	ErrorContestProjectIDMustBeGreaterThanZero = httperror.New(http.StatusBadRequest, "contest project id must be greater than zero")
	ErrorContestProjectNotFound                = httperror.New(http.StatusNotFound, "contest project not found")
	ErrorInternalServer                        = httperror.New(http.StatusInternalServerError, "internal server error")
)

type Input struct {
	UserID           int32
	ContestProjectID int32
}

type ProjectTime struct {
	project *project.Project
}

func (p *ProjectTime) TimeLeftToStartSubmit(currentTime time.Time) time.Duration {
	return p.project.TimeLeftToStartSubmit(currentTime)
}

func (p *ProjectTime) TimeLeftToSubmit(currentTime time.Time) time.Duration {
	return p.project.TimeLeftToSubmit(currentTime)
}

type Output struct {
	ProjectTime *ProjectTime
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
	if input.ContestProjectID <= 0 {
		return ErrorContestProjectIDMustBeGreaterThanZero
	}

	return nil
}

func (u *usecase) Execute(input *Input) (*Output, error) {
	u.logger.Infof("Starting to stream user project time", "projectId", input.ContestProjectID)

	if err := u.validate(input); err != nil {
		u.logger.Errorf("Failed to validate input", "projectId", input.ContestProjectID, "error", err)
		return nil, err
	}

	if err := u.contestProjectQuery.CheckUserCanAccessProject(input.ContestProjectID, input.UserID); err != nil {
		u.logger.Errorf("User cannot access project time stream", "userId", input.UserID, "projectId", input.ContestProjectID, "error", err)
		return nil, err
	}

	project, err := u.repository.FindByID(input.ContestProjectID)
	if err != nil {
		u.logger.Errorf("Failed to find project", "projectId", input.ContestProjectID, "error", err)
		return nil, err
	}

	if project == nil {
		u.logger.Errorf("Project not found", "projectId", input.ContestProjectID)
		return nil, ErrorContestProjectNotFound
	}

	u.logger.Infof("Successfully started streaming user project time", "projectId", input.ContestProjectID)
	return &Output{
		ProjectTime: &ProjectTime{
			project: project,
		},
	}, nil
}
