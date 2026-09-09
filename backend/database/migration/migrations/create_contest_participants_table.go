package migrations

import (
	"judge-backend/internal/app/shared/model"

	"gorm.io/gorm"
)

type ContestParticipantMigrator struct{}

func (m ContestParticipantMigrator) Up(db *gorm.DB) error {
	return db.AutoMigrate(&model.ContestParticipant{})
}

func (m ContestParticipantMigrator) Down(db *gorm.DB) error {
	return db.Migrator().DropTable(&model.ContestParticipant{})
}
