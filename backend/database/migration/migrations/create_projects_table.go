package migrations

import (
	"judge-backend/internal/app/shared/model"

	"gorm.io/gorm"
)

type ProjectMigrator struct{}

func (m ProjectMigrator) Up(db *gorm.DB) error {
	return db.AutoMigrate(&model.Project{})
}

func (m ProjectMigrator) Down(db *gorm.DB) error {
	return db.Migrator().DropTable(&model.Project{})
}
