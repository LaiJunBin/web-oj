package imageservice

import (
	"bytes"
	"errors"
	sharedservice "judge-backend/internal/app/shared/service"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/image"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/repository"
	buildservice "judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/service/build_service"
	imageservice "judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/service/image_service"
)

var (
	ErrorImageStatusCanNotChangeToBuilding = errors.New("image status can't change to building")
	ErrorImageNotFound                     = errors.New("image not found")
)

type Service struct {
	imageRepository repository.ImageRepository
	buildService    buildservice.BuildService
	fss             sharedservice.FileSystemService
	zipService      sharedservice.ZipService
}

func NewService(
	imageRepository repository.ImageRepository,
	buildService buildservice.BuildService,
	fss sharedservice.FileSystemService,
	zipService sharedservice.ZipService,
) imageservice.ImageService {

	return &Service{
		imageRepository: imageRepository,
		buildService:    buildService,
		fss:             fss,
		zipService:      zipService,
	}
}

func (s *Service) SetupProject(image *image.Image, zipProject *bytes.Buffer) error {
	// Clearing the tag's directory first is what makes a re-setup replace the
	// project instead of merging into it. Swallowing this error unpacked the
	// new project ON TOP of the old one, producing an image that builds from
	// a silent mix of both — so it stops the setup. RemoveAll returns nil for
	// a path that never existed, so a first-time setup is unaffected.
	if err := s.CleanupProject(image); err != nil {
		return err
	}
	if err := s.fss.MkdirAll(image.Project().StoragePath()); err != nil {
		return err
	}

	if err := s.zipService.UnzipToPath(zipProject, image.Project().StoragePath(), true); err != nil {
		return err
	}

	return nil
}

func (s *Service) CleanupProject(image *image.Image) error {
	return s.fss.RemoveAll(image.Project().StoragePath())
}

func (s *Service) ListProjectFiles(image *image.Image) ([]string, error) {
	return s.fss.ListFileRecursive(image.Project().StoragePath())
}
