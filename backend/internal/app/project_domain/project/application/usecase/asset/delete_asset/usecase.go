package deleteasset

import (
	"errors"
	"judge-backend/internal/app/project_domain/project/domain/repository"
	"judge-backend/internal/app/project_domain/project/domain/service"
	"log"
)

var (
	ErrorInvalidInput   = errors.New("invalid input")
	ErrorAssetNotFound  = errors.New("asset not found")
	ErrorInternalServer = errors.New("internal server error")
)

type Input struct {
	ID string
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
	if input.ID == "" {
		return ErrorInvalidInput
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

	if err := u.repository.Delete(asset); err != nil {
		return nil, ErrorInternalServer
	}

	if err := u.service.DeleteFile(input.ID); err != nil {
		log.Printf("failed to delete file %s: %v", input.ID, err)
	}

	return &Output{}, nil
}
