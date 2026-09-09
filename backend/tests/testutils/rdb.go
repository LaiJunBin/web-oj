package testutils

import (
	"context"
	"judge-backend/database"
	"judge-backend/internal/config"

	"github.com/redis/go-redis/v9"
)

var rdb = database.NewRedisDB(config.NewFeatureConfig())

func GetRedisClient() *redis.Client {
	return rdb
}

func ResetRDB(rdb *redis.Client) {
	rdb.FlushAll(context.Background())
}
