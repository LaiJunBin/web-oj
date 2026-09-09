package ignoresubmission

import (
	"judge-backend/internal/app/shared/types/httperror"
	"judge-backend/internal/app/contest_domain/activity/domain/repository"
	"net/http"
)

var (
	ErrorRequiredSubmissionID = httperror.New(http.StatusBadRequest, "Required submission ID")
	ErrorSubmissionNotFound   = httperror.New(http.StatusNotFound, "Submission not found")
	ErrorInternalServer       = httperror.New(http.StatusInternalServerError, "Internal server error")
)

type Input struct {
	SubmissionID string
}

type Output struct {
}

type Usecase interface {
	Execute(input *Input) (*Output, error)
}

type usecase struct {
	repository repository.SubmissionRepository
}

func NewUsecase(
	repository repository.SubmissionRepository,
) Usecase {
	return &usecase{
		repository: repository,
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

	submission, err := u.repository.FindByID(input.SubmissionID)
	if err != nil {
		return nil, err
	}

	if submission == nil {
		return nil, ErrorSubmissionNotFound
	}

	submission.MarkAsIgnored()
	if err := u.repository.Save(submission); err != nil {
		return nil, ErrorInternalServer
	}

	return &Output{}, nil
}
