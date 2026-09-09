package factory

import (
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/build"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/build/valueobject/batch"
	buildstatus "judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/build/valueobject/status"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/image"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/repository"
)

type BuildFactory interface {
	Create(image *image.Image) (*build.Build, error)
}

type BuildFactoryImpl struct {
	repo repository.BuildRepository
}

func NewBuildFactory(
	repo repository.BuildRepository,
) BuildFactory {
	return &BuildFactoryImpl{
		repo: repo,
	}
}

func (f *BuildFactoryImpl) Create(image *image.Image) (*build.Build, error) {
	maxBatch, err := f.repo.GetMaxBatch(image.Tag().String())
	if err != nil {
		return nil, err
	}

	batch, err := batch.New(maxBatch + 1)
	if err != nil {
		return nil, err
	}

	config := &build.BuildConfig{
		Tag:    image.Tag(),
		Batch:  batch,
		Status: buildstatus.PENDING,
	}

	build := build.New(config)
	return build, nil
}
