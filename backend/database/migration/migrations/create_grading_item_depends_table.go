package migrations

import (
	"judge-backend/internal/app/shared/model"

	"gorm.io/gorm"
)

type GradingItemDependMigrator struct{}

func (m GradingItemDependMigrator) Up(db *gorm.DB) error {
	return db.AutoMigrate(&model.GradingItemDepend{})
}

func (m GradingItemDependMigrator) Down(db *gorm.DB) error {
	return db.Migrator().DropTable(&model.GradingItemDepend{})
}
