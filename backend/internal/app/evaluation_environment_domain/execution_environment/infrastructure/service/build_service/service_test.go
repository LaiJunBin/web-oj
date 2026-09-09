package buildservice_test

import (
	"archive/tar"
	"bytes"
	sharedmocks "judge-backend/internal/app/shared/mocks"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/build"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/build/valueobject/batch"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/image"
	domainservice "judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/service/build_service"
	buildservice "judge-backend/internal/app/evaluation_environment_domain/execution_environment/infrastructure/service/build_service"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/mocks"
	"judge-backend/internal/projectpath"
	"testing"

	"github.com/stretchr/testify/mock"
)

type buildServiceMocks struct {
	eventBus     *sharedmocks.EventBus
	queue        *mocks.BuildQueue
	fss          *sharedmocks.FileSystemService
	buildFactory *mocks.BuildFactory
	buildRepo    *mocks.BuildRepository
}

func newBuildServiceMocks(t *testing.T) *buildServiceMocks {
	return &buildServiceMocks{
		eventBus:     sharedmocks.NewEventBus(t),
		queue:        mocks.NewBuildQueue(t),
		fss:          sharedmocks.NewFileSystemService(t),
		buildFactory: mocks.NewBuildFactory(t),
		buildRepo:    mocks.NewBuildRepository(t),
	}
}

func newBuildService(t *testing.T) (*buildservice.Service, *buildServiceMocks) {
	mocks := newBuildServiceMocks(t)
	service := buildservice.NewService(mocks.eventBus, mocks.queue, mocks.fss, mocks.buildFactory, mocks.buildRepo).(*buildservice.Service)
	return service, mocks
}

func createMockTarBuffer() *bytes.Buffer {
	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)

	var files = []struct {
		Name, Body string
	}{
		{"Dockerfile", "FROM nginx:latest\n"},
		{"index.html", "Hello, World!"},
	}

	for _, file := range files {
		hdr := &tar.Header{
			Name: file.Name,
			Mode: 0600,
			Size: int64(len(file.Body)),
		}

		err := tw.WriteHeader(hdr)
		if err != nil {
			panic(err)
		}

		_, err = tw.Write([]byte(file.Body))
		if err != nil {
			panic(err)
		}
	}

	err := tw.Close()
	if err != nil {
		panic(err)
	}

	return buf
}

func TestBuildService_EnqueueBuildJob(t *testing.T) {
	t.Run("Fail if image project not set", func(t *testing.T) {
		service, mocks := newBuildService(t)
		mockImage := image.MustNew(&image.Config{
			Tag:         "nginx:latest",
			Description: "description1",
			StoragePath: projectpath.Root + "/build_service_test",
			MountPath:   "/var/www/html",
		})

		mocks.fss.On("Exists", mockImage.Project().StoragePath()).Return(false)
		err := service.EnqueueBuildJob(mockImage, &domainservice.BuildOptions{})
		if err == nil {
			t.Fatalf("Expected error, got nil")
		}
	})

	t.Run("Success", func(t *testing.T) {
		service, mocks := newBuildService(t)
		mockImage := image.MustNew(&image.Config{
			Tag:         "nginx:latest",
			Description: "description1",
			StoragePath: projectpath.Root + "/build_service_test",
			MountPath:   "/var/www/html",
		})
		mockBuild := build.New(&build.BuildConfig{
			Tag:   mockImage.Tag(),
			Batch: batch.MustNew(1),
		})

		mocks.fss.On("Exists", mockImage.Project().StoragePath()).Return(true)
		mocks.buildFactory.On("Create", mockImage).Return(mockBuild, nil)
		mocks.buildRepo.On("SavePendingExclusive", mockBuild).Return(nil)
		mocks.queue.On("Enqueue", mock.Anything)

		err := service.EnqueueBuildJob(mockImage, &domainservice.BuildOptions{})
		if err != nil {
			t.Fatalf("Expected nil, got %s", err)
		}
	})
}
