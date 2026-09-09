package listusercontests

import (
	"judge-backend/internal/app/contest_domain/contest/application/query"
	"judge-backend/internal/app/contest_domain/contest/application/query/dto"
	sharedtypes "judge-backend/internal/app/shared/types"
)

type Input struct {
	UserID int32
}

type Output struct {
	Contests []dto.Contest
}

type Usecase interface {
	Execute(input *Input) (*Output, error)
}

type usecase struct {
	query  query.ContestQuery
	logger sharedtypes.Logger
}

func NewUsecase(
	query query.ContestQuery,
	logger sharedtypes.Logger,
) Usecase {
	return &usecase{
		query:  query,
		logger: logger,
	}
}

func (u *usecase) Execute(input *Input) (*Output, error) {
	u.logger.Infof("Starting to list user contests", "userId", input.UserID)

	contests, err := u.query.ListUserContests(input.UserID)
	if err != nil {
		u.logger.Errorf("Failed to list user contests from database", "userId", input.UserID, "error", err)
		return nil, err
	}

	u.logger.Infof("Successfully listed user contests", "userId", input.UserID)
	return &Output{
		Contests: contests,
	}, nil
}
