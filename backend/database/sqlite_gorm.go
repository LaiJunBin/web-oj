package database

import (
	"judge-backend/database/migration"
	"judge-backend/database/seeder"
	"judge-backend/internal/config"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func NewSqliteGormDB(config *config.Config) *gorm.DB {
	// The sqlite driver is an IN-MEMORY demonstration database (note the
	// mode=memory DSN) that is wiped on every restart and auto-seeded with the
	// well-known demo accounts (admin, demo01–05). Running it in production
	// would ship those fixed credentials and lose all data on restart, so refuse
	// rather than let a misconfiguration do that silently.
	if config.IsProduction() {
		panic("rdbDriver 'sqlite' is the in-memory demonstration database and must not be used in production; configure a persistent database (e.g. mysql)")
	}

	db, err := gorm.Open(sqlite.Open("file:memdb1?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	migration.RunMigrate(db)
	seeder.RunSeed(db, config)

	return db
}
