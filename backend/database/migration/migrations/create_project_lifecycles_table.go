package migrations

import (
	"judge-backend/internal/app/shared/model"

	"gorm.io/gorm"
)

type ProjectLifecycleMigrator struct{}

func (m ProjectLifecycleMigrator) Up(db *gorm.DB) error {
	return db.AutoMigrate(&model.Lifecycle{})
}

func (m ProjectLifecycleMigrator) Down(db *gorm.DB) error {
	return db.Migrator().DropTable(&model.Lifecycle{})
}
