package migrations

import (
	"judge-backend/internal/app/shared/model"

	"gorm.io/gorm"
)

type GradingItemMigrator struct{}

func (m GradingItemMigrator) Up(db *gorm.DB) error {
	return db.AutoMigrate(&model.GradingItem{})
}

func (m GradingItemMigrator) Down(db *gorm.DB) error {
	return db.Migrator().DropTable(&model.GradingItem{})
}
