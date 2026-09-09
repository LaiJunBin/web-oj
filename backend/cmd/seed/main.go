package main

import (
	"fmt"
	"judge-backend/app/provider"
	"judge-backend/database/seeder"
	"judge-backend/internal/config"

	"go.uber.org/fx"
	"gorm.io/gorm"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in f", r)
		}
	}()

	provides := append([]interface{}{
		config.NewConfig,
	}, provider.Databases...)

	app := fx.New(
		fx.Provide(
			provides...,
		),
		fx.Invoke(
			invokeSeeders,
		),
	)
	app.Run()
}

func invokeSeeders(db *gorm.DB, cfg *config.Config, shutdowner fx.Shutdowner) {
	seeder.RunSeed(db, cfg)
	shutdowner.Shutdown()
}
