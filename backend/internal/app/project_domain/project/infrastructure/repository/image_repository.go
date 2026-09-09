package repository

import (
	"judge-backend/internal/app/shared/model"
	"judge-backend/internal/app/project_domain/project/domain/repository"

	"gorm.io/gorm"
)

type ImageRepository struct {
	db *gorm.DB
}

func NewImageRepository(
	db *gorm.DB,
) repository.ImageRepository {
	return &ImageRepository{
		db: db,
	}
}

func (r *ImageRepository) Exists(tag string) (bool, error) {
	var count int64
	if err := r.db.Model(&model.Image{}).Where("tag = ?", tag).Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}
