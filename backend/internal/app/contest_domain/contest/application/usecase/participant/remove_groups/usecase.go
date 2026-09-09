package removegroups

import (
	"judge-backend/internal/app/contest_domain/contest/domain/repository"
)

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
	if err := u.repo.RemoveGroups(input.ContestID, input.GroupIDs); err != nil {
		return nil, err
	}
	return &Output{}, nil
}
