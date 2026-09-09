package listcontest

import (
	"judge-backend/internal/app/contest_domain/contest/application/query"
	"judge-backend/internal/app/contest_domain/contest/application/query/dto"
	"judge-backend/internal/app/shared/types/httperror"
	"net/http"
)

var (
	ErrorInternalServer = httperror.New(http.StatusInternalServerError, "internal server error")
)

type Input struct {
}

type Output struct {
	Contests []*dto.AdminContest
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

func (u *usecase) Execute(input *Input) (*Output, error) {
	contests, err := u.query.ListContest()
	if err != nil {
		return nil, ErrorInternalServer
	}

	return &Output{
		Contests: contests,
	}, nil
}
