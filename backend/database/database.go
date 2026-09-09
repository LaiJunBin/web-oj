package database

import (
	"judge-backend/internal/config"

	"gorm.io/gorm"
)

func DatabaseProvider(config *config.Config) *gorm.DB {
	switch config.Server.RdbDriver {
	case "mysql":
		return NewMysqlGormDB(config)
	case "sqlite":
		return NewSqliteGormDB(config)
	default:
		panic("invalid RdbDriver")
	}
}
