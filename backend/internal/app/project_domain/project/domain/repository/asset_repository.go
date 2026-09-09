package repository

import "judge-backend/internal/app/project_domain/project/domain/entity/asset"

type AssetRepository interface {
	FindByID(id string) (*asset.Asset, error)
	Save(asset *asset.Asset) error
	Delete(asset *asset.Asset) error
}
