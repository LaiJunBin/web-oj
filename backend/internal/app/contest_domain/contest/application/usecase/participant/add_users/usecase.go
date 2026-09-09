package addusers

import (
	"judge-backend/internal/app/contest_domain/contest/domain/repository"
	"judge-backend/internal/app/shared/types/httperror"
	"net/http"
)

var ErrorUserIDsRequired = httperror.New(http.StatusBadRequest, "userIds is required")

type Input struct {
	ContestID int32
	UserIDs   []int
}

type Output struct{}

type Usecase interface {
	Execute(input *Input) (*Output, error)
}

type usecase struct {
	repo repository.ParticipantRepository
}

func New(repo repository.ParticipantRepository) Usecase {
	return &usecase{repo: repo}
}

func (u *usecase) Execute(input *Input) (*Output, error) {
	if len(input.UserIDs) == 0 {
		return nil, ErrorUserIDsRequired
	}
	if err := u.repo.AddUsers(input.ContestID, input.UserIDs); err != nil {
		return nil, err
	}
	return &Output{}, nil
}
