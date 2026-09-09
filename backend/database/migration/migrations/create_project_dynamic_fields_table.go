package migrations

import (
	"judge-backend/internal/app/shared/model"

	"gorm.io/gorm"
)

type ProjectDynamicFieldMigrator struct{}

func (m ProjectDynamicFieldMigrator) Up(db *gorm.DB) error {
	return db.AutoMigrate(&model.DynamicField{})
}

func (m ProjectDynamicFieldMigrator) Down(db *gorm.DB) error {
	return db.Migrator().DropTable(&model.DynamicField{})
}
