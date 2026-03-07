package container

import (
	"fmt"

	docker "github.com/fsouza/go-dockerclient"
)

func NewClient() (*docker.Client, error) {
	cli, err := docker.NewClientFromEnv()
	if err != nil {
		return nil, fmt.Errorf("connecting to Docker: %w\nIs Docker running?", err)
	}
	return cli, nil
}
