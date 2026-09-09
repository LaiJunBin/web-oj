package migrations

import (
	"judge-backend/internal/app/shared/model"

	"gorm.io/gorm"
)

type BuildMigrator struct{}

func (m BuildMigrator) Up(db *gorm.DB) error {
	return db.AutoMigrate(&model.Build{})
}

func (m BuildMigrator) Down(db *gorm.DB) error {
	return db.Migrator().DropTable(&model.Build{})
}
