package database

import (
	"fmt"
	"judge-backend/internal/config"

	"github.com/gin-contrib/sessions/redis"
)

func NewRedisStore(config *config.Config) redis.Store {
	store, err := redis.NewStore(
		10,
		"tcp",
		fmt.Sprintf("%s:%d", config.Database.Redis.Host, config.Database.Redis.Port), config.Database.Redis.Password,
		[]byte(config.Server.StoreKey),
	)
	if err != nil {
		panic(err)
	}

	return store
}
