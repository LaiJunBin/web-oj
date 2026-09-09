package getuserprojectscoreboard

import (
	"judge-backend/internal/app/contest_domain/contest/application/query"
	"judge-backend/internal/app/contest_domain/contest/application/query/dto"
	"judge-backend/internal/app/contest_domain/contest/domain/repository"
	sharedtypes "judge-backend/internal/app/shared/types"
)

type Input struct {
	UserID           int32
	ContestProjectID int32
}

type Output struct {
	Scoreboard *dto.Scoreboard
}

type Usecase interface {
	Execute(input *Input) (*Output, error)
}

type usecase struct {
	repository          repository.ProjectRepository
	query               query.ProjectQuery
	contestProjectQuery query.ContestProjectQuery
	logger              sharedtypes.Logger
}

func NewUsecase(
	repository repository.ProjectRepository,
	query query.ProjectQuery,
	contestProjectQuery query.ContestProjectQuery,
	logger sharedtypes.Logger,
) Usecase {
	return &usecase{
		repository:          repository,
		query:               query,
		contestProjectQuery: contestProjectQuery,
		logger:              logger,
	}
}

func (u *usecase) Execute(input *Input) (*Output, error) {
	u.logger.Infof("Starting to get user project scoreboard", "userId", input.UserID, "projectId", input.ContestProjectID)

	if err := u.contestProjectQuery.CheckUserCanAccessProject(input.ContestProjectID, input.UserID); err != nil {
		u.logger.Errorf("User cannot access project scoreboard", "userId", input.UserID, "projectId", input.ContestProjectID, "error", err)
		return nil, err
	}

	project, err := u.repository.FindByID(input.ContestProjectID)
	if err != nil {
		u.logger.Errorf("Failed to find project by ID", "projectId", input.ContestProjectID, "error", err)
		return nil, err
	}

	scoreboard, err := u.query.GetProjectScoreboardByID(project.ID())
	if err != nil {
		u.logger.Errorf("Failed to get project scoreboard from database", "projectId", input.ContestProjectID, "error", err)
		return nil, err
	}

	u.logger.Infof("Successfully got user project scoreboard", "userId", input.UserID, "projectId", input.ContestProjectID)
	return &Output{
		Scoreboard: scoreboard,
	}, nil
}
