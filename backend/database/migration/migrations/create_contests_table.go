package migrations

import (
	"judge-backend/internal/app/shared/model"

	"gorm.io/gorm"
)

type ContestMigrator struct{}

func (m ContestMigrator) Up(db *gorm.DB) error {
	return db.AutoMigrate(&model.Contest{})
}

func (m ContestMigrator) Down(db *gorm.DB) error {
	return db.Migrator().DropTable(&model.Contest{})
}
