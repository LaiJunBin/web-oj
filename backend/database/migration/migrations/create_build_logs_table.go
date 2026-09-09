package migrations

import (
	"judge-backend/internal/app/shared/model"

	"gorm.io/gorm"
)

type BuildLogMigrator struct{}

func (m BuildLogMigrator) Up(db *gorm.DB) error {
	return db.AutoMigrate(&model.BuildLog{})
}

func (m BuildLogMigrator) Down(db *gorm.DB) error {
	return db.Migrator().DropTable(&model.BuildLog{})
}
