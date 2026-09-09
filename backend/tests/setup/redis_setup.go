package setup

import (
	"fmt"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func (s *setup) setupRedis() error {
	image := "redis:alpine"
	fmt.Println("Setting up Redis container with image:", image)

	request := testcontainers.ContainerRequest{
		Image:        image,
		ExposedPorts: []string{"6379"},
		WaitingFor:   wait.ForLog("Ready to accept connections"),
	}

	redisContainer, err := testcontainers.GenericContainer(s.ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: request,
		Started:          true,
	})

	if err != nil {
		return err
	}

	host, err := redisContainer.Host(s.ctx)

	if err != nil {
		return err
	}

	port, err := redisContainer.MappedPort(s.ctx, "6379")

	if err != nil {
		return err
	}

	s.config.Database.Redis.Host = host
	s.config.Database.Redis.Port = port.Int()

	s.containers[image] = redisContainer
	return nil
}
