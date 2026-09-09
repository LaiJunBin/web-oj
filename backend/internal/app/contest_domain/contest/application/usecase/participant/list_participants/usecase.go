package listparticipants

import (
	"judge-backend/internal/app/contest_domain/contest/domain/repository"
)

type Input struct {
	ContestID int32
}

type Output struct {
	UserIDs  []int
	GroupIDs []int
}

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
	userIDs, err := u.repo.GetUserIDs(input.ContestID)
	if err != nil {
		return nil, err
	}

	groupIDs, err := u.repo.GetGroupIDs(input.ContestID)
	if err != nil {
		return nil, err
	}

	return &Output{
		UserIDs:  userIDs,
		GroupIDs: groupIDs,
	}, nil
}
