package adduserstogroup

import (
	"judge-backend/internal/app/identity_domain/user/domain/repository"
)

type Input struct {
	GroupID int
	UserIDs []int
}

type Output struct {
}

type Usecase interface {
	Execute(input *Input) (*Output, error)
}

type usecase struct {
	repo repository.GroupRepository
}

func New(repo repository.GroupRepository) Usecase {
	return &usecase{repo: repo}
}

func (u *usecase) Execute(input *Input) (*Output, error) {
	if err := u.repo.AddUsersToGroup(input.GroupID, input.UserIDs); err != nil {
		return nil, err
	}
	return &Output{}, nil
}
