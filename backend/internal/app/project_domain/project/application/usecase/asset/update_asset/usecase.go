package updateasset

import (
	"bytes"
	"errors"
	"judge-backend/internal/app/project_domain/project/domain/repository"
	"judge-backend/internal/app/project_domain/project/domain/service"
	"log"
)

var (
	ErrorInvalidInput    = errors.New("invalid input")
	ErrorInvalidFileName = errors.New("invalid file name")
	ErrorInvalidFile     = errors.New("invalid file")
	ErrorAssetNotFound   = errors.New("asset not found")
	ErrorInternalServer  = errors.New("internal server error")
)

type Input struct {
	ID        string
	FileName  *string
	File      *bytes.Buffer
	IsEnabled *bool
}

type Output struct {
}

type Usecase interface {
	Execute(input *Input) (*Output, error)
}

type usecase struct {
	repository repository.AssetRepository
	service    service.AssetService
}

func NewUsecase(
	repository repository.AssetRepository,
	service service.AssetService,
) Usecase {
	return &usecase{
		repository: repository,
		service:    service,
	}
}

func (u *usecase) validate(input *Input) error {
	if input.FileName == nil && input.File == nil && input.IsEnabled == nil {
		return ErrorInvalidInput
	}

	if input.FileName != nil && *input.FileName == "" {
		return ErrorInvalidFileName
	}

	if input.File != nil && input.File.Len() == 0 {
		return ErrorInvalidFile
	}

	return nil
}

func (u *usecase) Execute(input *Input) (*Output, error) {
	if err := u.validate(input); err != nil {
		return nil, err
	}

	asset, err := u.repository.FindByID(input.ID)
	if err != nil {
		return nil, err
	}

	if asset == nil {
		return nil, ErrorAssetNotFound
	}

	if input.FileName != nil {
		if err := asset.SetDisplayName(*input.FileName); err != nil {
			return nil, ErrorInvalidFileName
		}
	}

	if input.IsEnabled != nil {
		asset.SetIsEnabled(*input.IsEnabled)
	}

	if input.File != nil {
		newAsset, err := u.service.UploadFile(asset.ProjectID().Value(), asset.DisplayName().String(), input.File)
		if err != nil {
			return nil, ErrorInternalServer
		}

		if err := u.repository.Delete(asset); err != nil {
			return nil, ErrorInternalServer
		}

		if err := u.service.DeleteFile(asset.ID().String()); err != nil {
			log.Printf("failed to delete file %s: %v", asset.ID().String(), err)
		}

		asset = newAsset
	}

	if err := u.repository.Save(asset); err != nil {
		return nil, ErrorInternalServer
	}

	return &Output{}, nil
}
