package setup

import (
	"fmt"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func (s *setup) setupMySQL() error {
	image := "mysql:8.0.26"
	fmt.Println("Setting up MySQL container with image:", image)

	env := map[string]string{
		"MYSQL_ROOT_PASSWORD": "root",
		"MYSQL_DATABASE":      "judge",
	}

	request := testcontainers.ContainerRequest{
		Image:        image,
		ExposedPorts: []string{"3306"},
		Env:          env,
		// 3 minutes had no margin. Reaching this log line means initializing a fresh
		// datadir, starting a temporary server, running the setup SQL, shutting it
		// down and starting the real one — measured at 176s on a machine busy with
		// other containers, against a 180s budget, so the suite failed before it ran
		// a single scenario. This is a wait-for-log, not a sleep: a fast machine
		// still proceeds as soon as the line appears and pays nothing for the
		// headroom.
		WaitingFor: wait.ForLog("port: 3306  MySQL Community Server - GPL").WithStartupTimeout(6 * time.Minute),
	}

	mysqlContainer, err := testcontainers.GenericContainer(s.ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: request,
		Started:          true,
	})

	if err != nil {
		return err
	}

	host, err := mysqlContainer.Host(s.ctx)

	if err != nil {
		return err
	}

	port, err := mysqlContainer.MappedPort(s.ctx, "3306")

	if err != nil {
		return err
	}

	s.config.Server.RdbDriver = "mysql"
	s.config.Database.Mysql.Host = host
	s.config.Database.Mysql.Port = port.Int()
	s.config.Database.Mysql.Username = "root"
	s.config.Database.Mysql.Password = env["MYSQL_ROOT_PASSWORD"]
	s.config.Database.Mysql.DBName = env["MYSQL_DATABASE"]

	s.containers[image] = mysqlContainer
	return nil
}
