package provider

import (
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/provider"
	"judge-backend/internal/config"
)

type StoragePathProvider struct {
	config *config.Config
}

func NewStoragePathProvider(
	config *config.Config,
) provider.StoragePathProvider {
	return &StoragePathProvider{
		config: config,
	}
}

func (p *StoragePathProvider) GetStoragePath() string {
	return p.config.Server.StoragePath
}
