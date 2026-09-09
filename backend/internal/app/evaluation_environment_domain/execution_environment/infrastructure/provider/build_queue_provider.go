package provider

import (
	sharedtypes "judge-backend/internal/app/shared/types"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/infrastructure/queue"
	"judge-backend/internal/config"
)

func NewBuildQueueProvider(eventbus sharedtypes.EventBus, config *config.Config) queue.BuildQueue {
	return queue.NewBuildQueue(eventbus, 2)
}
