package getsubmissionsbycontestproject

import (
	"judge-backend/internal/app/shared/types/httperror"
	"judge-backend/internal/app/contest_domain/activity/application/query"
	"judge-backend/internal/app/contest_domain/activity/application/query/dto"
	"net/http"
)

var (
	ErrorInvalidContestProjectID = httperror.New(http.StatusBadRequest, "Invalid contest project ID")
)

type Input struct {
	ContestProjectID int32
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
	if input.ContestProjectID <= 0 {
		return ErrorInvalidContestProjectID
	}

	return nil
}

func (u *usecase) Execute(input *Input) (*Output, error) {
	if err := u.validate(input); err != nil {
		return nil, err
	}

	submissions, err := u.query.GetSubmissionsByContestProjectID(input.ContestProjectID)
	if err != nil {
		return nil, err
	}

	return &Output{
		Submissions: submissions,
	}, nil
}
