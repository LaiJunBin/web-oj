package getimage

import (
	"errors"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/repository"
	"time"
)

var (
	ErrorTagIsRequired  = errors.New("tag is required")
	ErrorImageNotFound  = errors.New("image not found")
	ErrorInternalServer = errors.New("internal server error")
)

type Input struct {
	Tag string
}

type Output struct {
	Tag         string
	MountPath   string
	Description string
	IsReady     bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Usecase interface {
	Execute(input *Input) (*Output, error)
}

type usecase struct {
	imageRepository repository.ImageRepository
}

func NewUsecase(
	imageRepository repository.ImageRepository,
) Usecase {
	return &usecase{
		imageRepository: imageRepository,
	}
}

func (u *usecase) validate(input *Input) error {
	if input.Tag == "" {
		return ErrorTagIsRequired
	}

	return nil
}

func (u *usecase) Execute(input *Input) (*Output, error) {
	if err := u.validate(input); err != nil {
		return nil, err
	}

	image, err := u.imageRepository.FindByTag(input.Tag)
	if err != nil {
		return nil, ErrorInternalServer
	}

	if image == nil {
		return nil, ErrorImageNotFound
	}

	return &Output{
		Tag:         image.Tag().String(),
		MountPath:   image.MountPath().String(),
		Description: image.Description(),
		IsReady:     image.IsReady(),
		CreatedAt:   image.CreatedAt(),
		UpdatedAt:   image.UpdatedAt(),
	}, nil
}
