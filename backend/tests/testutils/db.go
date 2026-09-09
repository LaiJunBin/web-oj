package testutils

import (
	"judge-backend/database"
	"judge-backend/database/migration"
	"judge-backend/database/seeder"
	"judge-backend/internal/config"

	"gorm.io/gorm"
)

var cfg = config.NewFeatureConfig()
var db = database.DatabaseProvider(cfg)

func GetDB() *gorm.DB {
	return db
}

func RunMigration(db *gorm.DB) {
	migration.RunMigrate(db)
}

func RunSeeder(db *gorm.DB) {
	seeder.RunSeed(db, cfg)
}

func RunMigrationsAndSeeders(db *gorm.DB) {
	RunMigration(db)
	RunSeeder(db)
}

func ResetDB(db *gorm.DB) {
	migration.RunRollback(db)
}
