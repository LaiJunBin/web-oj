package getpublicprojectscoreboard

import (
	"judge-backend/internal/app/contest_domain/contest/application/query"
	"judge-backend/internal/app/contest_domain/contest/application/query/dto"
	"judge-backend/internal/app/contest_domain/contest/domain/repository"
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
	Scoreboard *dto.Scoreboard
}

type Usecase interface {
	Execute(input *Input) (*Output, error)
}

type usecase struct {
	repository repository.ProjectRepository
	query      query.ProjectQuery
	logger     sharedtypes.Logger
}

func NewUsecase(
	repository repository.ProjectRepository,
	query query.ProjectQuery,
	logger sharedtypes.Logger,
) Usecase {
	return &usecase{
		repository: repository,
		query:      query,
		logger:     logger,
	}
}

func (u *usecase) Execute(input *Input) (*Output, error) {
	u.logger.Infof("Starting to get public project scoreboard", "projectId", input.ContestProjectID)

	project, err := u.repository.FindByID(input.ContestProjectID)
	if err != nil {
		u.logger.Errorf("Failed to find project", "projectId", input.ContestProjectID, "error", err)
		return nil, err
	}

	if !project.IsPublic() {
		u.logger.Errorf("Project is not public", "projectId", input.ContestProjectID)
		return nil, ErrorProjectNotFound
	}

	scoreboard, err := u.query.GetProjectScoreboardByID(project.ID())
	if err != nil {
		u.logger.Errorf("Failed to get project scoreboard", "projectId", input.ContestProjectID, "error", err)
		return nil, err
	}

	u.logger.Infof("Successfully got public project scoreboard", "projectId", input.ContestProjectID)
	return &Output{
		Scoreboard: scoreboard,
	}, nil
}
