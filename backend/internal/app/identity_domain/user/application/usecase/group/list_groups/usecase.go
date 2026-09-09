package listgroups

import (
	"judge-backend/internal/app/identity_domain/user/domain/entity/group"
	"judge-backend/internal/app/identity_domain/user/domain/repository"
)

type Input struct {
}

type Output struct {
	Groups []*group.Group
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
	groups, err := u.repo.FindAll()
	if err != nil {
		return nil, err
	}

	return &Output{Groups: groups}, nil
}
