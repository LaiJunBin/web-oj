package query

import (
	"errors"
	"judge-backend/internal/app/project_domain/project/application/query/dto"
)

var (
	ErrorAssetNotFound = errors.New("asset not found")
)

type AssetQuery interface {
	GetAsset(id string) (*dto.Asset, error)
	DownloadAsset(fileName string) (*dto.DownloadFile, error)
}
