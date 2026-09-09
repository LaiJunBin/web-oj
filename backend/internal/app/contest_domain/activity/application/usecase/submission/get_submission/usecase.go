package getsubmission

import (
	"judge-backend/internal/app/shared/types/httperror"
	"judge-backend/internal/app/contest_domain/activity/application/query"
	"judge-backend/internal/app/contest_domain/activity/application/query/dto"
	"net/http"
)

var (
	ErrorRequiredSubmissionID = httperror.New(http.StatusBadRequest, "Required submission ID")
)

type Input struct {
	SubmissionID string
}

type Output struct {
	Submission *dto.SubmissionDetail
}

type Usecase interface {
	Execute(input *Input) (*Output, error)
}

type usecase struct {
	query query.SubmissionQuery
}

func NewUsecase(
	query query.SubmissionQuery,
) Usecase {
	return &usecase{
		query: query,
	}
}

func (u *usecase) validate(input *Input) error {
	if input.SubmissionID == "" {
		return ErrorRequiredSubmissionID
	}

	return nil
}

func (u *usecase) Execute(input *Input) (*Output, error) {
	if err := u.validate(input); err != nil {
		return nil, err
	}

	submission, err := u.query.GetSubmissionByID(input.SubmissionID)
	if err != nil {
		return nil, err
	}

	return &Output{
		Submission: submission,
	}, nil
}
