package deletecontest

import (
	"judge-backend/internal/app/contest_domain/contest/domain/repository"
	"judge-backend/internal/app/shared/types/httperror"
	"net/http"
)

var (
	ErrorInvalidID = httperror.New(http.StatusBadRequest, "invalid id, id must be greater than 0")
)

type Input struct {
	ID int32
}

type Output struct {
}

type Usecase interface {
	Execute(input *Input) (*Output, error)
}

type usecase struct {
	repository repository.ContestRepository
}

func NewUsecase(
	repository repository.ContestRepository,
) Usecase {
	return &usecase{
		repository: repository,
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

	contest, err := u.repository.FindByID(input.ID)
	if err != nil {
		return nil, err
	}

	if err := u.repository.Delete(contest); err != nil {
		return nil, err
	}

	return nil, nil
}
