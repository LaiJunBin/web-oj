package service

import (
	"bytes"
	"fmt"
	"judge-backend/internal/app/project_domain/project/domain/entity/asset"
	"judge-backend/internal/app/project_domain/project/domain/entity/project"
	"judge-backend/internal/app/project_domain/project/domain/service"
	"judge-backend/internal/config"
	"os"
	"time"

	"github.com/google/uuid"
)

type AssetService struct {
	storagePath string
}

// NewAssetService returns an error instead of panicking when the storage
// directory cannot be created: fx surfaces a constructor error as a clean
// startup failure naming this constructor, which beats a bare panic trace.
func NewAssetService(config *config.Config) (service.AssetService, error) {
	storagePath := fmt.Sprintf("%s/assets", config.Server.StoragePath)
	// MkdirAll is a no-op for an existing directory, so no Stat pre-check.
	// The old Stat+IsNotExist guard also silently swallowed ENOTDIR (a file
	// where the directory belongs), deferring the failure to the first
	// upload instead of failing startup honestly.
	if err := os.MkdirAll(storagePath, 0755); err != nil {
		return nil, fmt.Errorf("create asset storage dir %s: %w", storagePath, err)
	}

	return &AssetService{
		storagePath: storagePath,
	}, nil
}

func (s *AssetService) UploadFile(projectID int32, fileName string, file *bytes.Buffer) (*asset.Asset, error) {
	id := uuid.New().String()
	filePath := fmt.Sprintf("%s/%s", s.storagePath, id)

	f, err := os.Create(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	if _, err := f.Write(file.Bytes()); err != nil {
		return nil, err
	}

	return asset.New(&asset.AssetConfig{
		ID:          id,
		ProjectID:   project.ID(projectID),
		DisplayName: fileName,
		Size:        int64(file.Len()),
		LastMod:     time.Now(),
	})

}

func (s *AssetService) DeleteFile(fileName string) error {
	filePath := fmt.Sprintf("%s/%s", s.storagePath, fileName)
	return os.Remove(filePath)
}
