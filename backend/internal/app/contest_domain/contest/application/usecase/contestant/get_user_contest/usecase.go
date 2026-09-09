package getusercontest

import (
	"judge-backend/internal/app/contest_domain/contest/application/query"
	"judge-backend/internal/app/contest_domain/contest/application/query/dto"
	sharedtypes "judge-backend/internal/app/shared/types"
)

type Input struct {
	UserID    int32
	ContestID int32
}

type Output struct {
	Contest dto.ContestDetail
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
	u.logger.Infof("Starting to get user contest", "userId", input.UserID, "contestId", input.ContestID)

	contest, err := u.query.GetUserContestByID(input.UserID, input.ContestID)
	if err != nil {
		u.logger.Errorf("Failed to get user contest from database", "userId", input.UserID, "contestId", input.ContestID, "error", err)
		return nil, err
	}

	u.logger.Infof("Successfully got user contest", "userId", input.UserID, "contestId", input.ContestID)
	return &Output{
		Contest: contest,
	}, nil
}
