package migrations

import (
	"judge-backend/internal/app/shared/model"

	"gorm.io/gorm"
)

type ClarificationMigrator struct{}

func (m ClarificationMigrator) Up(db *gorm.DB) error {
	return db.AutoMigrate(&model.Clarification{})
}

func (m ClarificationMigrator) Down(db *gorm.DB) error {
	return db.Migrator().DropTable(&model.Clarification{})
}
