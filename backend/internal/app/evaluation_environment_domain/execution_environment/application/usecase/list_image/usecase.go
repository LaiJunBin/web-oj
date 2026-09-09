package listimage

import (
	"errors"
	imagequery "judge-backend/internal/app/evaluation_environment_domain/execution_environment/application/service/query/image_query"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/application/service/query/image_query/models/imageinfo"
)

var (
	ErrorInternalServer = errors.New("internal server error")
)

type Input struct {
}

type Output struct {
	Images []imageinfo.ImageInfo
}

type Usecase interface {
	Execute(input *Input) (*Output, error)
}

type usecase struct {
	service imagequery.ImageQueryService
}

func NewUsecase(
	service imagequery.ImageQueryService,
) Usecase {
	return &usecase{
		service: service,
	}
}

func (u *usecase) Execute(input *Input) (*Output, error) {
	images, err := u.service.GetImages()
	if err != nil {
		return nil, ErrorInternalServer
	}

	return &Output{
		Images: images,
	}, nil
}
