package listgradingprojectfiles

import (
	"errors"
	gradingproject "judge-backend/internal/app/project_domain/project/application/service/grading_project"
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
	Files []string
}

type Usecase interface {
	Execute(input *Input) (*Output, error)
}

type usecase struct {
	repository repository.ProjectRepository
	service    gradingproject.GradingProjectService
}

func NewUsecase(
	repository repository.ProjectRepository,
	service gradingproject.GradingProjectService,
) Usecase {
	return &usecase{
		repository: repository,
		service:    service,
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

	if exists, err := u.repository.Exists(int(input.ProjectID)); err != nil {
		return nil, ErrorInternalServer
	} else if !exists {
		return nil, ErrorProjectNotFound
	}

	files, err := u.service.ListFile(input.ProjectID)
	if err != nil {
		return nil, ErrorInternalServer
	}

	return &Output{
		Files: files,
	}, nil
}
