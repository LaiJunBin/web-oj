package provider

import (
	"judge-backend/internal/app/contest_domain/contest/domain/provider"
	sharedtypes "judge-backend/internal/app/shared/types"
	"judge-backend/internal/config"
)

type StoragePathProvider struct {
	config *config.Config
	logger sharedtypes.Logger
}

func NewStoragePathProvider(
	config *config.Config,
	logger sharedtypes.Logger,
) provider.StoragePathProvider {
	return &StoragePathProvider{
		config: config,
		logger: logger,
	}
}

func (p *StoragePathProvider) GetStoragePath() string {
	p.logger.Infof("Getting storage path from config")
	path := p.config.Server.StoragePath
	p.logger.Infof("Retrieved storage path", "path", path)
	return path
}
