package database

import (
	"context"
	"fmt"
	"judge-backend/internal/config"

	"github.com/redis/go-redis/v9"
)

func NewRedisDB(config *config.Config) *redis.Client {
	ctx := context.Background()
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.Database.Redis.Host, config.Database.Redis.Port),
		Password: config.Database.Redis.Password,
		DB:       config.Database.Redis.DB,
		PoolSize: 1000,
	})

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		panic(err)
	}

	return rdb
}
