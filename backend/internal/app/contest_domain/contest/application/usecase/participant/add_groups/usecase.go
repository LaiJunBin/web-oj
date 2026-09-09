package addgroups

import (
	"judge-backend/internal/app/contest_domain/contest/domain/repository"
	"judge-backend/internal/app/shared/types/httperror"
	"net/http"
)

var ErrorGroupIDsRequired = httperror.New(http.StatusBadRequest, "groupIds is required")

type Input struct {
	ContestID int32
	GroupIDs  []int
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
	if len(input.GroupIDs) == 0 {
		return nil, ErrorGroupIDsRequired
	}
	if err := u.repo.AddGroups(input.ContestID, input.GroupIDs); err != nil {
		return nil, err
	}
	return &Output{}, nil
}
