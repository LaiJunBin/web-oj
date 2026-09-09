package putgradingproject

import (
	"bytes"
	"errors"
	gradingproject "judge-backend/internal/app/project_domain/project/application/service/grading_project"
	"judge-backend/internal/app/project_domain/project/domain/repository"
)

var (
	ErrorInvalidInput       = errors.New("invalid input")
	ErrorInvalidProjectID   = errors.New("invalid project id, must be greater than 0")
	ErrorZipProjectRequired = errors.New("zip project required")
	ErrorProjectNotFound    = errors.New("project not found")
	ErrorInternalServer     = errors.New("internal server error")
)

type Input struct {
	ProjectID  int32
	ZipProject *bytes.Buffer
}

type Output struct {
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

	if input.ZipProject == nil {
		return ErrorZipProjectRequired
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

	if err := u.service.UploadGradingProject(input.ProjectID, input.ZipProject); err != nil {
		return nil, ErrorInternalServer
	}

	return &Output{}, nil
}
