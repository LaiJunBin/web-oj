package migrations

import (
	"judge-backend/internal/app/shared/model"

	"gorm.io/gorm"
)

type UserMigrator struct{}

func (m UserMigrator) Up(db *gorm.DB) error {
	return db.AutoMigrate(&model.User{})
}

func (m UserMigrator) Down(db *gorm.DB) error {
	return db.Migrator().DropTable(&model.User{})
}
