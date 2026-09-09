package provider

import (
	"fmt"
	"judge-backend/internal/config"
	"log"

	"github.com/docker/docker/client"
)

func getLocalDockerClient() *client.Client {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}
	return cli
}

func getRemoteDockerClient(config *config.Config) *client.Client {
	host := fmt.Sprintf("%s://%s:%d", config.Docker.Protocol, config.Docker.Host, config.Docker.Port)

	// The docker API is root-equivalent on the host. A remote connection without
	// mutual TLS would let anyone on the path issue daemon commands, so refuse to
	// start rather than connect in the clear (fail closed) — unless the operator
	// has explicitly opted in for a trusted, isolated private network.
	ca, cert, key := config.Docker.TLSCACert, config.Docker.TLSCert, config.Docker.TLSKey
	if ca == "" || cert == "" || key == "" {
		if !config.Docker.AllowInsecureRemote {
			panic("docker mode 'remote' requires TLS: set docker.tlsCACert, docker.tlsCert and docker.tlsKey (or, only on a trusted private network, set docker.allowInsecureRemote: true)")
		}
		log.Printf("WARNING: connecting to remote docker daemon at %s over plain TCP with no TLS; the docker API is root-equivalent, so only do this on a trusted, isolated network (docker.allowInsecureRemote is on)", host)
		cli, err := client.NewClientWithOpts(client.WithHost(host))
		if err != nil {
			panic(err)
		}
		return cli
	}

	cli, err := client.NewClientWithOpts(
		client.WithHost(host),
		client.WithTLSClientConfig(ca, cert, key),
	)
	if err != nil {
		panic(err)
	}

	return cli
}

func NewDockerClient(config *config.Config) client.APIClient {
	if config.Docker.Mode == "local" {
		return getLocalDockerClient()
	} else if config.Docker.Mode == "remote" {
		return getRemoteDockerClient(config)
	} else {
		panic("invalid docker mode")
	}
}
