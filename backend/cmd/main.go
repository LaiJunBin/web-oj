package main

import (
	"judge-backend/app"
	"judge-backend/app/provider"
	"judge-backend/internal/config"

	"go.uber.org/fx"
)

func main() {
	provides := append([]interface{}{
		config.NewConfig,
	}, provider.GetProvides()...)
	invokes := append(provider.GetInvokes(), app.SetupHandler)

	fx.New(
		fx.Provide(
			provides...,
		),
		fx.Invoke(
			invokes...,
		),
	).Run()
}
