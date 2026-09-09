package listgradingitem

import (
	"errors"
	"judge-backend/internal/app/project_domain/project/application/query"
	"judge-backend/internal/app/project_domain/project/application/query/dto"
	"judge-backend/internal/app/project_domain/project/domain/repository"
)

var (
	ErrorInvalidProjectID = errors.New("invalid project id, must be greater than 0")
	ErrorProjectNotFound  = errors.New("project not found")
	ErrorInternalServer   = errors.New("internal server error")
)

type Input struct {
	ProjectID int32
}

type Output struct {
	GradingItems []*dto.GradingItem
}

type Usecase interface {
	Execute(input *Input) (*Output, error)
}

type usecase struct {
	query       query.GradingItemQuery
	projectRepo repository.ProjectRepository
}

func NewUsecase(
	query query.GradingItemQuery,
	projectRepo repository.ProjectRepository,
) Usecase {
	return &usecase{
		query:       query,
		projectRepo: projectRepo,
	}
}

func (u *usecase) validate(input *Input) error {
	if input.ProjectID <= 0 {
		return ErrorInvalidProjectID
	}

	return nil
}

func (u *usecase) Execute(input *Input) (*Output, error) {
	if err := u.validate(input); err != nil {
		return nil, err
	}

	if exists, err := u.projectRepo.Exists(int(input.ProjectID)); err != nil {
		return nil, ErrorInternalServer
	} else if !exists {
		return nil, ErrorProjectNotFound
	}

	gradingItems, err := u.query.ListGradingItem(input.ProjectID)
	if err != nil {
		return nil, ErrorInternalServer
	}

	return &Output{GradingItems: gradingItems}, nil
}
