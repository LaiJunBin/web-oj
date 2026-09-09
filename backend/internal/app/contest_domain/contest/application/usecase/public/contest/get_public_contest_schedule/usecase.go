package getpubliccontestschedule

import (
	"judge-backend/internal/app/contest_domain/contest/application/query"
	"judge-backend/internal/app/contest_domain/contest/application/query/dto"
	sharedtypes "judge-backend/internal/app/shared/types"
	"judge-backend/internal/app/shared/types/httperror"
	"net/http"
)

var (
	ErrorInvalidContestID = httperror.New(http.StatusBadRequest, "invalid contest id")
	ErrorNotFoundContest  = httperror.New(http.StatusNotFound, "contest not found")
)

type Input struct {
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
	u.logger.Infof("Starting to get public contest schedule", "contestId", input.ContestID)

	if err := u.validate(input); err != nil {
		u.logger.Errorf("Failed to validate input", "contestId", input.ContestID, "error", err)
		return nil, err
	}

	schedule, err := u.query.GetPublicContestScheduleByID(input.ContestID)
	if err != nil {
		u.logger.Errorf("Failed to get public contest schedule", "contestId", input.ContestID, "error", err)
		return nil, err
	}

	if !schedule.IsPublic {
		u.logger.Errorf("Contest is not public", "contestId", input.ContestID)
		return nil, ErrorNotFoundContest
	}

	u.logger.Infof("Successfully got public contest schedule", "contestId", input.ContestID)
	return &Output{
		ContestSchedule: schedule,
	}, nil
}
