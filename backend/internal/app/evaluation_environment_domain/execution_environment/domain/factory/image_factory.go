package factory

import (
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/image"
	imagetag "judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/image/valueobject/tag"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/provider"
	"judge-backend/internal/projectpath"
	"path"
)

type ImageFactory struct {
	storagePath string
}

func NewImageFactory(storageProvider provider.StoragePathProvider) *ImageFactory {
	return &ImageFactory{
		storagePath: path.Join(projectpath.Root, storageProvider.GetStoragePath()),
	}
}

func (f *ImageFactory) CreateImage(tag string, mountPath string, description string) (*image.Image, error) {
	// The rename-on-delete namespace is reserved: a live image whose tag
	// literally contains the marker could collide with a dead row's renamed
	// tag, turning that row's delete into a unique-index failure. Every
	// creation path (create, import source/archive) funnels through this
	// factory; the check is here rather than in Tag.Validate() because
	// build-history hydration must keep ACCEPTING marker tags it reads back
	// off dead rows.
	if imagetag.ContainsDeletedMarker(tag) {
		return nil, imagetag.ErrInvalidTagFormat
	}
	image, err := image.New(&image.Config{
		Tag:         tag,
		MountPath:   mountPath,
		Description: description,
		StoragePath: f.storagePath,
	})
	if err != nil {
		return nil, err
	}

	return image, nil
}
