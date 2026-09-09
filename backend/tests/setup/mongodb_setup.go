package setup

import (
	"fmt"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func (s *setup) setupMongoDB() error {
	image := "mongo:4.4.18"
	fmt.Println("Setting up Mongo container with image:", image)

	request := testcontainers.ContainerRequest{
		Image:        image,
		ExposedPorts: []string{"27017"},
		Env: map[string]string{
			"MONGO_INITDB_ROOT_USERNAME": s.config.Database.Mongo.Username,
			"MONGO_INITDB_ROOT_PASSWORD": s.config.Database.Mongo.Password,
		},
		WaitingFor: wait.ForLog("Waiting for connections"),
	}

	mongoContainer, err := testcontainers.GenericContainer(s.ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: request,
		Started:          true,
	})

	if err != nil {
		return err
	}

	host, err := mongoContainer.Host(s.ctx)

	if err != nil {
		return err
	}

	port, err := mongoContainer.MappedPort(s.ctx, "27017")

	if err != nil {
		return err
	}

	s.config.Database.Mongo.Host = host
	s.config.Database.Mongo.Port = port.Int()

	s.containers[image] = mongoContainer
	return nil
}
