package migrations

import (
	"judge-backend/internal/app/shared/model"

	"gorm.io/gorm"
)

type ImageMigrator struct{}

func (m ImageMigrator) Up(db *gorm.DB) error {
	return db.AutoMigrate(&model.Image{})
}

func (m ImageMigrator) Down(db *gorm.DB) error {
	return db.Migrator().DropTable(&model.Image{})
}
