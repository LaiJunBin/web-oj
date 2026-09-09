package factory_test

import (
	"errors"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/image"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/factory"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/mocks"
	"testing"
)

type buildFactoryMocks struct {
	repo *mocks.BuildRepository
}

func newBuildFactory(t *testing.T) (*factory.BuildFactoryImpl, *buildFactoryMocks) {
	mocks := &buildFactoryMocks{
		repo: mocks.NewBuildRepository(t),
	}
	factory := factory.NewBuildFactory(mocks.repo).(*factory.BuildFactoryImpl)
	return factory, mocks
}

func TestBuildFactory(t *testing.T) {
	t.Run("should create a new build log, batch starts from 1", func(t *testing.T) {
		factory, mocks := newBuildFactory(t)
		tag := "test-nginx:latest"
		batch := 1
		mockImage := image.MustNew(&image.Config{
			Tag:         tag,
			Description: "test image",
			MountPath:   "/var/www/html",
		})
		mocks.repo.On("GetMaxBatch", tag).Return(0, nil)

		build, err := factory.Create(mockImage)
		if err != nil {
			t.Fatalf("Create(%s) should not return an error, but got %v", tag, err)
		}

		if build.Batch().Value() != batch {
			t.Errorf("New(%s) should create a new build log with batch %d, but got %d", tag, batch, build.Batch().Value())
		}
	})

	t.Run("should create a new build log, assuming there are existing logs batch is 1", func(t *testing.T) {
		factory, mocks := newBuildFactory(t)
		tag := "test-nginx:latest"
		expectBatch := 2
		mockImage := image.MustNew(&image.Config{
			Tag:         tag,
			Description: "test image",
			MountPath:   "/var/www/html",
		})
		mocks.repo.On("GetMaxBatch", tag).Return(1, nil)

		build, err := factory.Create(mockImage)
		if err != nil {
			t.Fatalf("Create(%s) should not return an error, but got %v", tag, err)
		}

		if build.Batch().Value() != expectBatch {
			t.Errorf("New(%s) should create a new build log with batch %d, but got %d", tag, expectBatch, build.Batch().Value())
		}
	})

	t.Run("should return an error when failed to get max batch", func(t *testing.T) {
		factory, mocks := newBuildFactory(t)
		tag := "test-nginx:latest"
		mockImage := image.MustNew(&image.Config{
			Tag:         tag,
			Description: "test image",
			MountPath:   "/var/www/html",
		})
		mocks.repo.On("GetMaxBatch", tag).Return(0, errors.New("failed to get max batch"))

		_, err := factory.Create(mockImage)
		if err == nil {
			t.Errorf("Create(%s) should return an error, but got nil", tag)
		}
	})
}
