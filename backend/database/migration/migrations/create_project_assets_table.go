package migrations

import (
	"judge-backend/internal/app/shared/model"

	"gorm.io/gorm"
)

type ProjectAssetMigrator struct{}

func (m ProjectAssetMigrator) Up(db *gorm.DB) error {
	return db.AutoMigrate(&model.Asset{})
}

func (m ProjectAssetMigrator) Down(db *gorm.DB) error {
	return db.Migrator().DropTable(&model.Asset{})
}
