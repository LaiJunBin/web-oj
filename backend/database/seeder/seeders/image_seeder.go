package seeders

import (
	"judge-backend/internal/app/shared/model"

	"gorm.io/gorm"
)

type ImageSeeder struct{}

func (s ImageSeeder) Seed(db *gorm.DB) error {
	data, err := LoadDemoData()
	if err != nil {
		return err
	}

	// isReady stays false in the dataset: the image still has to be built once
	// from the admin UI, which is the first step of the operating manual.
	return db.Create(&model.Image{
		Tag:         data.Image.Tag,
		MountPath:   data.Image.MountPath,
		Description: data.Image.Description,
		IsReady:     data.Image.IsReady,
	}).Error
}
