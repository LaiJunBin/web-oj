package uploadasset

import (
	"bytes"
	"errors"
	"judge-backend/internal/app/project_domain/project/domain/repository"
	"judge-backend/internal/app/project_domain/project/domain/service"
)

var (
	ErrorInvalidInput     = errors.New("invalid input")
	ErrorInvalidProjectID = errors.New("invalid project id, must be greater than 0")
	ErrorInvalidFileName  = errors.New("invalid file name")
	ErrorInvalidFile      = errors.New("invalid file")
	ErrorProjectNotFound  = errors.New("project not found")
	ErrorInternalServer   = errors.New("internal server error")
)

type Input struct {
	ProjectID int32
	FileName  string
	File      *bytes.Buffer
}

type Output struct {
	ID string
}

type Usecase interface {
	Execute(input *Input) (*Output, error)
}

type usecase struct {
	projectRepo repository.ProjectRepository
	repository  repository.AssetRepository
	service     service.AssetService
}

func NewUsecase(
	repository repository.AssetRepository,
	projectRepo repository.ProjectRepository,
	service service.AssetService,
) Usecase {
	return &usecase{
		repository:  repository,
		projectRepo: projectRepo,
		service:     service,
	}
}

func (u *usecase) validate(input *Input) error {
	if input.ProjectID <= 0 {
		return ErrorInvalidProjectID
	}

	if input.FileName == "" {
		return ErrorInvalidFileName
	}

	if input.File == nil {
		return ErrorInvalidFile
	}

	return nil
}

func (u *usecase) Execute(input *Input) (*Output, error) {
	if err := u.validate(input); err != nil {
		return nil, err
	}

	if exists, err := u.projectRepo.Exists(int(input.ProjectID)); err != nil {
		return nil, err
	} else if !exists {
		return nil, ErrorProjectNotFound
	}

	asset, err := u.service.UploadFile(input.ProjectID, input.FileName, input.File)
	if err != nil {
		return nil, ErrorInternalServer
	}

	if err := u.repository.Save(asset); err != nil {
		return nil, ErrorInternalServer
	}

	return &Output{
		ID: asset.ID().String(),
	}, nil
}
