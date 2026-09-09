package service

import (
	"bytes"
	"judge-backend/internal/app/project_domain/project/domain/entity/asset"
)

type AssetService interface {
	UploadFile(projectID int32, fileName string, file *bytes.Buffer) (*asset.Asset, error)
	DeleteFile(fileName string) error
}
