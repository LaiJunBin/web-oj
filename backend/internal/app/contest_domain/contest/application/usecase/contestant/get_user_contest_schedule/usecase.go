package getusercontestschedule

import (
	"judge-backend/internal/app/contest_domain/contest/application/query"
	"judge-backend/internal/app/contest_domain/contest/application/query/dto"
	sharedtypes "judge-backend/internal/app/shared/types"
	"judge-backend/internal/app/shared/types/httperror"
	"net/http"
)

var (
	ErrorInvalidContestID = httperror.New(http.StatusBadRequest, "invalid contest id")
)

type Input struct {
	UserID    int32
	ContestID int32
}

type Output struct {
	ContestSchedule dto.ContestSchedule
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

func (u *usecase) validate(input *Input) error {
	if input.ContestID <= 0 {
		return ErrorInvalidContestID
	}

	return nil
}

func (u *usecase) Execute(input *Input) (*Output, error) {
	u.logger.Infof("Starting to get user contest schedule", "contestId", input.ContestID)

	if err := u.validate(input); err != nil {
		u.logger.Errorf("Failed to validate input for getting user contest schedule", "contestId", input.ContestID, "error", err)
		return nil, err
	}

	schedule, err := u.query.GetUserContestScheduleByID(input.UserID, input.ContestID)
	if err != nil {
		u.logger.Errorf("Failed to get user contest schedule from database", "contestId", input.ContestID, "error", err)
		return nil, err
	}

	u.logger.Infof("Successfully got user contest schedule", "contestId", input.ContestID)
	return &Output{
		ContestSchedule: schedule,
	}, nil
}
