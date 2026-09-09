package getrejudgesubmissioncomparison

import (
	"judge-backend/internal/app/contest_domain/activity/application/query"
	"judge-backend/internal/app/contest_domain/activity/application/query/dto"
	"judge-backend/internal/app/shared/types/httperror"
	"net/http"
)

var (
	ErrRejudgeIDIsRequired    = httperror.New(http.StatusBadRequest, "rejudge_id is required")
	ErrSubmissionIDIsRequired = httperror.New(http.StatusBadRequest, "submission_id is required")
)

type Input struct {
	RejudgeID    string
	SubmissionID string
}

type Output struct {
	Comparison *dto.SubmissionComparisonDetail
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

	if input.SubmissionID == "" {
		return ErrSubmissionIDIsRequired
	}

	return nil
}

func (u *usecase) Execute(input *Input) (*Output, error) {
	if err := u.validate(input); err != nil {
		return nil, err
	}

	comparison, err := u.query.GetSubmissionComparison(input.RejudgeID, input.SubmissionID)
	if err != nil {
		return nil, err
	}

	return &Output{
		Comparison: comparison,
	}, nil
}
