package listuserprojectclarifications

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
	Clarifications []*dto.Clarification
}

type Usecase interface {
	Execute(input *Input) (*Output, error)
}

type usecase struct {
	query               query.ClarificationQuery
	contestProjectQuery query.ContestProjectQuery
	logger              sharedtypes.Logger
}

func NewUsecase(
	query query.ClarificationQuery,
	contestProjectQuery query.ContestProjectQuery,
	logger sharedtypes.Logger,
) Usecase {
	return &usecase{
		query:               query,
		contestProjectQuery: contestProjectQuery,
		logger:              logger,
	}
}

func (u *usecase) Execute(input *Input) (*Output, error) {
	u.logger.Infof("Starting to list user project clarifications", "userId", input.UserID, "projectId", input.ContestProjectID)

	if err := u.contestProjectQuery.CheckUserCanAccessProject(input.ContestProjectID, input.UserID); err != nil {
		u.logger.Errorf("User cannot access project clarifications", "userId", input.UserID, "projectId", input.ContestProjectID, "error", err)
		return nil, err
	}

	clarifications, err := u.query.ListUserClarificationsByProjectID(input.UserID, input.ContestProjectID)
	if err != nil {
		u.logger.Errorf("Failed to list user project clarifications", "userId", input.UserID, "projectId", input.ContestProjectID, "error", err)
		return nil, err
	}

	u.logger.Infof("Successfully listed user project clarifications", "userId", input.UserID, "projectId", input.ContestProjectID)
	return &Output{
		Clarifications: clarifications,
	}, nil
}
