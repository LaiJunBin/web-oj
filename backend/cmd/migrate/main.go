package main

import (
	"judge-backend/app/provider"
	"judge-backend/database/migration"
	"judge-backend/internal/config"

	"go.uber.org/fx"
	"gorm.io/gorm"
)

func main() {
	provides := append([]interface{}{
		config.NewConfig,
	}, provider.Databases...)

	app := fx.New(
		fx.Provide(
			provides...,
		),
		fx.Invoke(
			invokeMigrate,
		),
	)
	app.Run()
}

func invokeMigrate(db *gorm.DB, shutdowner fx.Shutdowner) {
	migration.RunMigrate(db)
	shutdowner.Shutdown()
}
