package getpubliccontest

import (
	"judge-backend/internal/app/contest_domain/contest/application/query"
	"judge-backend/internal/app/contest_domain/contest/application/query/dto"
	sharedtypes "judge-backend/internal/app/shared/types"
)

type Input struct {
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
	u.logger.Infof("Starting to get public contest", "contestId", input.ContestID)

	contest, err := u.query.GetPublicContestByID(input.ContestID)
	if err != nil {
		u.logger.Errorf("Failed to get public contest from database", "contestId", input.ContestID, "error", err)
		return nil, err
	}

	u.logger.Infof("Successfully got public contest", "contestId", input.ContestID)
	return &Output{
		Contest: contest,
	}, nil
}
