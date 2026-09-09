package getsubmissionsbyuserid

import (
	"judge-backend/internal/app/shared/types/httperror"
	"judge-backend/internal/app/contest_domain/activity/application/query"
	"judge-backend/internal/app/contest_domain/activity/application/query/dto"
	"net/http"
)

var (
	ErrorInvalidUserID = httperror.New(http.StatusBadRequest, "Invalid user ID")
)

type Input struct {
	UserID int32
}

type Output struct {
	Submissions []*dto.Submission
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
	if input.UserID <= 0 {
		return ErrorInvalidUserID
	}

	return nil
}

func (u *usecase) Execute(input *Input) (*Output, error) {
	if err := u.validate(input); err != nil {
		return nil, err
	}

	submissions, err := u.query.GetSubmissionsByUserID(input.UserID)
	if err != nil {
		return nil, err
	}

	return &Output{
		Submissions: submissions,
	}, nil
}
