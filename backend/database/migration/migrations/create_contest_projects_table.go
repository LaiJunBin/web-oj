package migrations

import (
	"judge-backend/internal/app/shared/model"

	"gorm.io/gorm"
)

type ContestProjectMigrator struct{}

func (m ContestProjectMigrator) Up(db *gorm.DB) error {
	return db.AutoMigrate(&model.ContestProject{})
}

func (m ContestProjectMigrator) Down(db *gorm.DB) error {
	return db.Migrator().DropTable(&model.ContestProject{})
}
