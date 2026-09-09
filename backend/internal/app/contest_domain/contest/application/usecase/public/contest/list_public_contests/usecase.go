package listpubliccontests

import (
	"judge-backend/internal/app/contest_domain/contest/application/query"
	"judge-backend/internal/app/contest_domain/contest/application/query/dto"
	sharedtypes "judge-backend/internal/app/shared/types"
)

type Input struct {
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
	u.logger.Infof("Starting to list public contests")

	contests, err := u.query.ListPublicContests()
	if err != nil {
		u.logger.Errorf("Failed to list public contests from database", "error", err)
		return nil, err
	}

	u.logger.Infof("Successfully listed public contests")
	return &Output{
		Contests: contests,
	}, nil
}
