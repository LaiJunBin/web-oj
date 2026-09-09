package getsubmissionrate

import (
	"judge-backend/internal/app/contest_domain/activity/application/query"
	"judge-backend/internal/app/contest_domain/activity/application/query/dto"
	"judge-backend/internal/app/shared/types/httperror"
	"net/http"
)

var (
	ErrorInvalidMinutes          = httperror.New(http.StatusBadRequest, "Minutes must be between 1 and 1440")
	ErrorInvalidContestProjectID = httperror.New(http.StatusBadRequest, "Invalid contest project ID")
)

const (
	// DefaultMinutes matches what the status page asks for when it does not say.
	DefaultMinutes int32 = 10

	// MaxMinutes caps the window at a day. The response carries one point per
	// minute whether or not anything happened in it, so an unbounded window is
	// an unbounded response — and a status page asking for a month of
	// minute-resolution data is a mistake, not a requirement.
	MaxMinutes int32 = 1440
)

type Input struct {
	// Minutes is nil when the caller did not specify a window.
	Minutes *int32

	// ContestProjectID is nil for the whole system.
	ContestProjectID *int32
}

type Output struct {
	Points []*dto.SubmissionRatePoint
}

type Usecase interface {
	Execute(input *Input) (*Output, error)
}

type usecase struct {
	query query.SystemStatusQuery
}

func NewUsecase(
	query query.SystemStatusQuery,
) Usecase {
	return &usecase{
		query: query,
	}
}

func (u *usecase) validate(input *Input) error {
	if input.Minutes != nil && (*input.Minutes <= 0 || *input.Minutes > MaxMinutes) {
		return ErrorInvalidMinutes
	}

	// Rejecting rather than ignoring: a non-positive id is a malformed request,
	// and silently widening it to "the whole system" would answer a question
	// the caller did not ask.
	if input.ContestProjectID != nil && *input.ContestProjectID <= 0 {
		return ErrorInvalidContestProjectID
	}

	return nil
}

func (u *usecase) Execute(input *Input) (*Output, error) {
	if err := u.validate(input); err != nil {
		return nil, err
	}

	minutes := DefaultMinutes
	if input.Minutes != nil {
		minutes = *input.Minutes
	}

	points, err := u.query.GetSubmissionRate(minutes, input.ContestProjectID)
	if err != nil {
		return nil, err
	}

	return &Output{
		Points: points,
	}, nil
}
