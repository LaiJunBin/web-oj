package shared

import (
	"judge-backend/internal/app/shared/infrastructure/eventbus"
	"judge-backend/internal/app/shared/infrastructure/logger"
	"judge-backend/internal/app/shared/infrastructure/metrics"
	sharedrepo "judge-backend/internal/app/shared/infrastructure/repository"
	filesystem "judge-backend/internal/app/shared/infrastructure/services/file_system"
	hashservice "judge-backend/internal/app/shared/infrastructure/services/hash_service"
	zipservice "judge-backend/internal/app/shared/infrastructure/services/zip_service"
	"judge-backend/internal/app/shared/types/provider"
	"judge-backend/internal/config"
)

// NewZipLimits is the single place the configured extraction limits become
// the zip service's Limits. Going through a provider keeps the zip service
// itself free of a dependency on the whole application config.
func NewZipLimits(cfg *config.Config) zipservice.Limits {
	return zipservice.Limits{
		MaxBytes:   cfg.Server.MaxUnzipBytes(),
		MaxEntries: cfg.Server.MaxUnzipEntries,
	}
}

func NewProvider() *provider.Provider {
	return provider.New(provider.Config{
		Infrastructures: []interface{}{
			filesystem.NewService,
			NewZipLimits,
			zipservice.NewService,
			hashservice.NewHashService,
			eventbus.NewEventBus,
			logger.NewZapLogger,
			sharedrepo.NewEventRepository,
		},
		Invokes: []interface{}{
			metrics.StartMetricsServer,
		},
	})
}
