package migrations

import (
	"judge-backend/internal/app/shared/model"

	"gorm.io/gorm"
)

type TestcaseMigrator struct{}

func (m TestcaseMigrator) Up(db *gorm.DB) error {
	return db.AutoMigrate(&model.Testcase{})
}

func (m TestcaseMigrator) Down(db *gorm.DB) error {
	return db.Migrator().DropTable(&model.Testcase{})
}
