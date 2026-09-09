package listjudgmenterrors

import (
	"judge-backend/internal/app/contest_domain/activity/application/query"
	"judge-backend/internal/app/contest_domain/activity/application/query/dto"
	"judge-backend/internal/app/shared/types/httperror"
	"net/http"
)

var (
	ErrorInvalidLimit = httperror.New(http.StatusBadRequest, "Limit must be between 1 and 200")
)

const (
	DefaultLimit int32 = 20

	// MaxLimit bounds the response. This list exists to answer "what just
	// broke"; a caller wanting the full error history wants a different,
	// paginated endpoint, not an unbounded one.
	MaxLimit int32 = 200
)

type Input struct {
	// Limit is nil when the caller did not specify one.
	Limit *int32
}

type Output struct {
	Errors []*dto.JudgmentError
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
	if input.Limit != nil && (*input.Limit <= 0 || *input.Limit > MaxLimit) {
		return ErrorInvalidLimit
	}

	return nil
}

func (u *usecase) Execute(input *Input) (*Output, error) {
	if err := u.validate(input); err != nil {
		return nil, err
	}

	limit := DefaultLimit
	if input.Limit != nil {
		limit = *input.Limit
	}

	errors, err := u.query.ListJudgmentErrors(limit)
	if err != nil {
		return nil, err
	}

	return &Output{
		Errors: errors,
	}, nil
}
