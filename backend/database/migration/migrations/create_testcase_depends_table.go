package migrations

import (
	"judge-backend/internal/app/shared/model"

	"gorm.io/gorm"
)

type TestcaseDependMigrator struct{}

func (m TestcaseDependMigrator) Up(db *gorm.DB) error {
	return db.AutoMigrate(&model.TestcaseDepend{})
}

func (m TestcaseDependMigrator) Down(db *gorm.DB) error {
	return db.Migrator().DropTable(&model.TestcaseDepend{})
}
