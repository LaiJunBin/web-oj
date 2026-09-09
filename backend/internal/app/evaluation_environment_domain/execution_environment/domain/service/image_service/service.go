package imageservice

import (
	"bytes"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/image"
)

type ImageService interface {
	SetupProject(image *image.Image, zipProject *bytes.Buffer) error
	CleanupProject(image *image.Image) error
	ListProjectFiles(image *image.Image) ([]string, error)
}
