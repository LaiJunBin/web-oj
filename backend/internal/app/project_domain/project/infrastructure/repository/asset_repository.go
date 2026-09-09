package repository

import (
	"judge-backend/internal/app/shared/model"
	"judge-backend/internal/app/project_domain/project/domain/entity/asset"
	"judge-backend/internal/app/project_domain/project/domain/entity/project"
	"judge-backend/internal/app/project_domain/project/domain/repository"

	"gorm.io/gorm"
)

type AssetRepository struct {
	db *gorm.DB
}

func NewAssetRepository(db *gorm.DB) repository.AssetRepository {
	return &AssetRepository{
		db: db,
	}
}

func (r *AssetRepository) toDomainAsset(a *model.Asset) (*asset.Asset, error) {
	domainAsset, err := asset.New(&asset.AssetConfig{
		ProjectID:   project.ID(a.ProjectID),
		DisplayName: a.Name,
		Size:        a.Size,
		Enabled:     a.IsEnabled,
		LastMod:     a.UpdatedAt,
	})
	if err != nil {
		return nil, err
	}

	if err := domainAsset.SetID(a.ID); err != nil {
		return nil, err
	}

	domainAsset.SetCreatedAt(a.CreatedAt)
	domainAsset.SetUpdatedAt(a.UpdatedAt)
	return domainAsset, nil
}

func (r *AssetRepository) FindByID(id string) (*asset.Asset, error) {
	asset := &model.Asset{}
	if err := r.db.First(asset, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}

		return nil, err
	}

	return r.toDomainAsset(asset)
}

func (r *AssetRepository) Save(asset *asset.Asset) error {
	a := &model.Asset{
		ID:          asset.ID().String(),
		Name:        asset.DisplayName().String(),
		Description: asset.DisplayName().String(),
		Size:        asset.Size().Bytes(),
		IsEnabled:   asset.IsEnabled(),
		ProjectID:   asset.ProjectID().Value(),
	}

	if !asset.CreatedAt().IsZero() {
		a.CreatedAt = asset.CreatedAt()
	}

	if err := r.db.Save(a).Error; err != nil {
		return err
	}

	return nil
}

func (r *AssetRepository) Delete(asset *asset.Asset) error {
	if err := r.db.Delete(&model.Asset{}, "id = ?", asset.ID().String()).Error; err != nil {
		return err
	}

	return nil
}
