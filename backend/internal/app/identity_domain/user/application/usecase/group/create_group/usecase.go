package creategroup

import (
	"judge-backend/internal/app/identity_domain/user/domain/entity/group"
	"judge-backend/internal/app/identity_domain/user/domain/repository"
)

type Input struct {
	Name        string
	Description string
}

type Output struct {
	Group *group.Group
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
	g, err := group.New(&group.GroupConfig{
		Name:        input.Name,
		Description: input.Description,
	})
	if err != nil {
		return nil, err
	}

	if err := u.repo.Save(g); err != nil {
		return nil, err
	}

	return &Output{Group: g}, nil
}
