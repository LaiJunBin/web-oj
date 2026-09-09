package updateclarification

import (
	"judge-backend/internal/app/shared/types/httperror"
	"judge-backend/internal/app/contest_domain/activity/domain/repository"
	"net/http"
)

var (
	ErrorIsAnsweredIsRequired  = httperror.New(http.StatusBadRequest, "isAnswered is required")
	ErrorClarificationNotFound = httperror.New(http.StatusNotFound, "clarification not found")
)

type Input struct {
	ID         int32
	IsAnswered *bool
}

type Output struct {
}

type Usecase interface {
	Execute(input *Input) (*Output, error)
}

type usecase struct {
	repository repository.ClarificationRepository
}

func NewUsecase(
	repository repository.ClarificationRepository,
) Usecase {
	return &usecase{
		repository: repository,
	}
}

func (u *usecase) validate(input *Input) error {
	if input.IsAnswered == nil {
		return ErrorIsAnsweredIsRequired
	}

	return nil
}

func (u *usecase) Execute(input *Input) (*Output, error) {
	if err := u.validate(input); err != nil {
		return nil, err
	}

	c, err := u.repository.FindByID(input.ID)
	if err != nil {
		return nil, err
	}

	if c == nil {
		return nil, ErrorClarificationNotFound
	}

	if c.IsAnswered() == *input.IsAnswered {
		return &Output{}, nil
	}

	c.SetIsAnswered(*input.IsAnswered)

	if err := u.repository.Save(c); err != nil {
		return nil, err
	}

	return &Output{}, nil
}
