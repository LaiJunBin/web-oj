package query

import (
	"errors"
	"fmt"
	"judge-backend/internal/app/shared/model"
	"judge-backend/internal/app/project_domain/project/application/query"
	"judge-backend/internal/app/project_domain/project/application/query/dto"
	"judge-backend/internal/config"
	"os"

	"gorm.io/gorm"
)

type AssetQuery struct {
	db          *gorm.DB
	storagePath string
}

func NewAssetQuery(
	db *gorm.DB,
	config *config.Config,
) query.AssetQuery {
	storagePath := fmt.Sprintf("%s/assets", config.Server.StoragePath)
	return &AssetQuery{
		db:          db,
		storagePath: storagePath,
	}
}

func (q *AssetQuery) GetAsset(id string) (*dto.Asset, error) {
	asset := &model.Asset{}
	if err := q.db.First(asset, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &dto.Asset{
		ID:           asset.ID,
		Name:         asset.Name,
		Size:         asset.Size,
		Enabled:      asset.IsEnabled,
		LastModified: asset.UpdatedAt,
	}, nil
}

func (q *AssetQuery) DownloadAsset(fileName string) (*dto.DownloadFile, error) {
	asset := &model.Asset{}
	if err := q.db.First(asset, "id = ?", fileName).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, query.ErrorAssetNotFound
		}
		return nil, err
	}

	path := fmt.Sprintf("%s/%s", q.storagePath, asset.ID)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, query.ErrorAssetNotFound
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	content := make([]byte, asset.Size)
	if _, err := file.Read(content); err != nil {
		return nil, err
	}

	return &dto.DownloadFile{
		Name:    asset.Name,
		Size:    asset.Size,
		Content: content,
	}, nil
}
