package provider

import (
	"judge-backend/internal/app/project_domain/project/domain/provider"
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
