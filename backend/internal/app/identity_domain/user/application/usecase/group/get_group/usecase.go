package getgroup

import (
	"judge-backend/internal/app/identity_domain/user/domain/entity/group"
	"judge-backend/internal/app/identity_domain/user/domain/repository"
	"judge-backend/internal/app/shared/types/httperror"
	"net/http"
)

var ErrorGroupNotFound = httperror.New(http.StatusNotFound, "group not found")

type Input struct {
	ID int
}

type Output struct {
	Group   *group.Group
	UserIDs []int
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
	g, err := u.repo.GetByID(input.ID)
	if err != nil {
		return nil, err
	}
	if g == nil {
		return nil, ErrorGroupNotFound
	}

	userIDs, err := u.repo.GetUserIDs(input.ID)
	if err != nil {
		return nil, err
	}

	return &Output{Group: g, UserIDs: userIDs}, nil
}
