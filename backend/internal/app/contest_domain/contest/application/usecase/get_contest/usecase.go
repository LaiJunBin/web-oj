package getcontest

import (
	"judge-backend/internal/app/contest_domain/contest/application/query"
	"judge-backend/internal/app/contest_domain/contest/application/query/dto"
	"judge-backend/internal/app/shared/types/httperror"
	"net/http"
)

var (
	ErrorInvalidID      = httperror.New(http.StatusBadRequest, "invalid id, must be greater than 0")
	ErrorInternalServer = httperror.New(http.StatusInternalServerError, "internal server error")
)

type Input struct {
	ID int32
}

type Output struct {
	Contest *dto.AdminContestDetail
}

type Usecase interface {
	Execute(input *Input) (*Output, error)
}

type usecase struct {
	query query.AdminContestQuery
}

func NewUsecase(
	query query.AdminContestQuery,
) Usecase {
	return &usecase{
		query: query,
	}
}

func (u *usecase) validate(input *Input) error {
	if input.ID <= 0 {
		return ErrorInvalidID
	}
	return nil
}

func (u *usecase) Execute(input *Input) (*Output, error) {
	if err := u.validate(input); err != nil {
		return nil, err
	}

	contest, err := u.query.GetContestByID(input.ID)
	if err != nil {
		return nil, err
	}

	return &Output{
		Contest: contest,
	}, nil
}
