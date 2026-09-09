package readgradingprojectfile

import (
	"errors"
	"judge-backend/internal/app/project_domain/project/application/service/base64"
	gradingproject "judge-backend/internal/app/project_domain/project/application/service/grading_project"
	"judge-backend/internal/app/project_domain/project/domain/repository"
)

var (
	ErrorInvalidProjectID          = errors.New("invalid project id, must be greater than 0")
	ErrorBase64EncodedPathRequired = errors.New("base64 encoded path required")
	ErrorProjectNotFound           = errors.New("project not found")
	ErrorInvalidBase64EncodedPath  = errors.New("invalid base64 encoded path")
	ErrorInternalServer            = errors.New("internal server error")
)

type Input struct {
	ProjectID         int32
	Base64EncodedPath string
}

type Output struct {
	Base64Content string
}

type Usecase interface {
	Execute(input *Input) (*Output, error)
}

type usecase struct {
	repository repository.ProjectRepository
	service    gradingproject.GradingProjectService
	base64     base64.Base64Service
}

func NewUsecase(
	repository repository.ProjectRepository,
	service gradingproject.GradingProjectService,
	base64 base64.Base64Service,
) Usecase {
	return &usecase{
		repository: repository,
		service:    service,
		base64:     base64,
	}
}

func (u *usecase) validate(input *Input) error {
	if input.ProjectID <= 0 {
		return ErrorInvalidProjectID
	}

	if input.Base64EncodedPath == "" {
		return ErrorBase64EncodedPathRequired
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

	decodedPath, err := u.base64.Decode(input.Base64EncodedPath)
	if err != nil {
		return nil, ErrorInvalidBase64EncodedPath
	}

	file, err := u.service.ReadFile(input.ProjectID, decodedPath)
	if err != nil {
		return nil, ErrorInternalServer
	}

	return &Output{
		Base64Content: u.base64.Encode(file.String()),
	}, nil
}
