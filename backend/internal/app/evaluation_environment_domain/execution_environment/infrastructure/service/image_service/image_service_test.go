package imageservice_test

import (
	"archive/zip"
	"bytes"
	"errors"
	"judge-backend/internal/app/shared/infrastructure/logger"
	sharedmocks "judge-backend/internal/app/shared/mocks"
	filesystem "judge-backend/internal/app/shared/infrastructure/services/file_system"
	zipservice "judge-backend/internal/app/shared/infrastructure/services/zip_service"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/image"
	imageservice "judge-backend/internal/app/evaluation_environment_domain/execution_environment/infrastructure/service/image_service"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/mocks"
	"judge-backend/internal/projectpath"
	"os"
	"testing"
)

type imageServiceMocks struct {
	imageRepository *mocks.ImageRepository
	buildService    *mocks.BuildService
}

func newMocks(t *testing.T) *imageServiceMocks {
	imageRepository := mocks.NewImageRepository(t)
	buildService := mocks.NewBuildService(t)
	return &imageServiceMocks{
		imageRepository: imageRepository,
		buildService:    buildService,
	}
}

func newImage(config *image.Config) *image.Image {
	image, err := image.New(config)
	if err != nil {
		panic(err)
	}

	return image
}

func createZipProject() *bytes.Buffer {
	buf := new(bytes.Buffer)

	w := zip.NewWriter(buf)

	var files = []struct {
		Name, Body string
	}{
		{"Dockerfile", "FROM nginx:latest\n"},
		{"index.html", "Hello, World!"},
	}
	for _, file := range files {
		f, err := w.Create(file.Name)
		if err != nil {
			panic(err)
		}
		_, err = f.Write([]byte(file.Body))
		if err != nil {
			panic(err)
		}
	}

	err := w.Close()
	if err != nil {
		panic(err)
	}

	return buf
}

func TestImageService_SetupProject(t *testing.T) {
	t.Run("error if when project file is not zip", func(t *testing.T) {
		mocks := newMocks(t)
		logger := logger.NewMockLogger()
		fss := filesystem.NewService(logger)
		zipservice := zipservice.NewService(logger, zipservice.Limits{})
		service := imageservice.NewService(mocks.imageRepository, mocks.buildService, fss, zipservice)
		mockImage := newImage(&image.Config{
			Tag:         "nginx:latest",
			Description: "description1",
			StoragePath: projectpath.Root + "/tmp/image_service_test",
			MountPath:   "/var/www/html",
		})
		os.MkdirAll(mockImage.Project().StoragePath(), 0755)
		defer os.RemoveAll(projectpath.Root + "/tmp/image_service_test")

		err := service.SetupProject(mockImage, &bytes.Buffer{})
		if err == nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	// Setting up a project starts by clearing whatever the tag's directory
	// already holds. If that clearing fails, the new project is unpacked ON
	// TOP of the leftovers and the resulting image is a silent mix of two —
	// so the failure has to stop the setup, not be swallowed.
	t.Run("error if the previous project cannot be cleared", func(t *testing.T) {
		mocks := newMocks(t)
		fss := sharedmocks.NewFileSystemService(t)
		zipservice := zipservice.NewService(logger.NewMockLogger(), zipservice.Limits{})
		service := imageservice.NewService(mocks.imageRepository, mocks.buildService, fss, zipservice)
		mockImage := newImage(&image.Config{
			Tag:         "nginx:latest",
			Description: "description1",
			StoragePath: projectpath.Root + "/tmp/image_service_test",
			MountPath:   "/var/www/html",
		})

		fss.On("RemoveAll", mockImage.Project().StoragePath()).
			Return(errors.New("permission denied")).Once()
		// No MkdirAll expectation: strict mocks fail the test if setup
		// carries on past the failed cleanup.

		if err := service.SetupProject(mockImage, createZipProject()); err == nil {
			t.Error("expected SetupProject to fail when the cleanup fails")
		}
	})

	t.Run("should setup project", func(t *testing.T) {
		mocks := newMocks(t)
		logger := logger.NewMockLogger()
		fss := filesystem.NewService(logger)
		zipservice := zipservice.NewService(logger, zipservice.Limits{})
		service := imageservice.NewService(mocks.imageRepository, mocks.buildService, fss, zipservice)
		mockImage := newImage(&image.Config{
			Tag:         "nginx:latest",
			Description: "description1",
			StoragePath: projectpath.Root + "/tmp/image_service_test",
			MountPath:   "/var/www/html",
		})
		os.MkdirAll(mockImage.Project().StoragePath(), 0755)
		defer os.RemoveAll(projectpath.Root + "/tmp/image_service_test")

		zipProject := createZipProject()
		err := service.SetupProject(mockImage, zipProject)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if !fss.Exists(mockImage.Project().StoragePath()) {
			t.Errorf("project not exists")
		}

		if !fss.Exists(mockImage.Project().StoragePath() + "/Dockerfile") {
			t.Errorf("Dockerfile not exists")
		}

		if !fss.Exists(mockImage.Project().StoragePath() + "/index.html") {
			t.Errorf("index.html not exists")
		}
	})
}

func TestImageService_ListProjectFiles(t *testing.T) {
	t.Run("should list project files", func(t *testing.T) {
		mocks := newMocks(t)
		logger := logger.NewMockLogger()
		fss := filesystem.NewService(logger)
		zipservice := zipservice.NewService(logger, zipservice.Limits{})
		service := imageservice.NewService(mocks.imageRepository, mocks.buildService, fss, zipservice)
		mockImage := newImage(&image.Config{
			Tag:         "nginx:latest",
			Description: "description1",
			StoragePath: projectpath.Root + "/tmp/image_service_test",
			MountPath:   "/var/www/html",
		})
		os.MkdirAll(mockImage.Project().StoragePath(), 0755)
		defer os.RemoveAll(projectpath.Root + "/tmp/image_service_test")
		os.Create(mockImage.Project().StoragePath() + "/file1.txt")
		os.Create(mockImage.Project().StoragePath() + "/file2.txt")

		files, err := service.ListProjectFiles(mockImage)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(files) != 2 {
			t.Errorf("unexpected files count: %v", len(files))
		}
	})
}
