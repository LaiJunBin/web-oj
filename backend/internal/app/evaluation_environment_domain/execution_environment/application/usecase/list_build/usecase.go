package listbuild

import (
	"errors"
	buildquery "judge-backend/internal/app/evaluation_environment_domain/execution_environment/application/service/query/build_query"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/application/service/query/build_query/models/buildinfo"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/repository"
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
	Builds []buildinfo.BuildInfo
}

type Usecase interface {
	Execute(input *Input) (*Output, error)
}

type usecase struct {
	imageRepository   repository.ImageRepository
	buildQueryService buildquery.BuildQueryService
}

func NewUsecase(
	imageRepository repository.ImageRepository,
	buildQueryService buildquery.BuildQueryService,
) Usecase {
	return &usecase{
		imageRepository:   imageRepository,
		buildQueryService: buildQueryService,
	}
}

func (uc *usecase) validate(input *Input) error {
	if input.Tag == "" {
		return ErrorTagIsRequired
	}

	return nil
}

func (uc *usecase) Execute(input *Input) (*Output, error) {
	if err := uc.validate(input); err != nil {
		return nil, err
	}

	image, err := uc.imageRepository.FindByTag(input.Tag)
	if err != nil {
		return nil, ErrorInternalServer
	}

	if image == nil {
		return nil, ErrorImageNotFound
	}

	builds, err := uc.buildQueryService.GetBuilds(input.Tag)
	if err != nil {
		return nil, ErrorInternalServer
	}

	return &Output{
		Builds: builds,
	}, nil
}
