package migrations

import (
	"judge-backend/internal/app/shared/model"

	"gorm.io/gorm"
)

type GroupMigrator struct{}

func (m GroupMigrator) Up(db *gorm.DB) error {
	return db.AutoMigrate(&model.Group{}, &model.UserGroup{})
}

func (m GroupMigrator) Down(db *gorm.DB) error {
	return db.Migrator().DropTable(&model.Group{}, &model.UserGroup{})
}
