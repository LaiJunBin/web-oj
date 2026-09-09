package getrejudge

import (
	"judge-backend/internal/app/shared/types/httperror"
	"judge-backend/internal/app/contest_domain/activity/application/query"
	"judge-backend/internal/app/contest_domain/activity/application/query/dto"
	"net/http"
)

var (
	ErrRejudgeIDIsRequired = httperror.New(http.StatusBadRequest, "rejudge_id is required")
)

type Input struct {
	RejudgeID string
}

type Output struct {
	Rejudge *dto.Rejudge
}

type Usecase interface {
	Execute(input *Input) (*Output, error)
}

type usecase struct {
	query query.RejudgeQuery
}

func NewUsecase(
	query query.RejudgeQuery,
) Usecase {
	return &usecase{
		query: query,
	}
}

func (u *usecase) validate(input *Input) error {
	if input.RejudgeID == "" {
		return ErrRejudgeIDIsRequired
	}

	return nil
}

func (u *usecase) Execute(input *Input) (*Output, error) {
	if err := u.validate(input); err != nil {
		return nil, err
	}

	rejudge, err := u.query.GetRejudgeByID(input.RejudgeID)
	if err != nil {
		return nil, err
	}

	return &Output{
		Rejudge: rejudge,
	}, nil
}
