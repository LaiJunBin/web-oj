package getclarification

import (
	"judge-backend/internal/app/shared/types/httperror"
	"judge-backend/internal/app/contest_domain/activity/application/query"
	"judge-backend/internal/app/contest_domain/activity/application/query/dto"
	"net/http"
)

var (
	ErrorClarificationNotFound = httperror.New(http.StatusNotFound, "clarification not found")
)

type Input struct {
	ClarificationID int32
}

type Output struct {
	Clarification *dto.ClarificationDetail
}

type Usecase interface {
	Execute(input *Input) (*Output, error)
}

type usecase struct {
	query query.ClarificationQuery
}

func NewUsecase(
	query query.ClarificationQuery,
) Usecase {
	return &usecase{
		query: query,
	}
}

func (u *usecase) Execute(input *Input) (*Output, error) {
	clarification, err := u.query.GetClarificationByID(input.ClarificationID)
	if err != nil {
		return nil, err
	}

	if clarification == nil {
		return nil, ErrorClarificationNotFound
	}

	return &Output{
		Clarification: clarification,
	}, nil
}
