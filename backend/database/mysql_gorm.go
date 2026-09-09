package database

import (
	"fmt"
	"judge-backend/internal/config"
	"net/url"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func NewMysqlGormDB(config *config.Config) *gorm.DB {
	loc := "Asia/Taipei"
	loc = url.QueryEscape(loc)

	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=true&loc=%s",
		config.Database.Mysql.Username,
		config.Database.Mysql.Password,
		config.Database.Mysql.Host,
		config.Database.Mysql.Port,
		config.Database.Mysql.DBName,
		loc,
	)

	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN: dsn,
	}))

	if err != nil {
		panic(err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		panic(err)
	}

	sqlDB.SetMaxOpenConns(config.Database.Mysql.MaxOpenConns)
	sqlDB.SetMaxIdleConns(config.Database.Mysql.MaxIdleConns)

	// Bound connection reuse so the pool never hands out a connection MySQL's
	// wait_timeout already closed server-side (config default: 5 minutes; an
	// explicit <= 0 disables the cap).
	if minutes := *config.Database.Mysql.ConnMaxLifetimeMinutes; minutes > 0 {
		sqlDB.SetConnMaxLifetime(time.Duration(minutes) * time.Minute)
	} else {
		sqlDB.SetConnMaxLifetime(0)
	}

	return db
}
