package container

import (
	"fmt"

	docker "github.com/fsouza/go-dockerclient"

	"github.com/weiyong1024/clawfleet/internal/config"
)

// EnsureNetwork creates the "clawfleet-net" container network if it does not
// already exist.
func EnsureNetwork(cli *docker.Client) error {
	networks, err := cli.ListNetworks()
	if err != nil {
		return fmt.Errorf("listing networks: %w", err)
	}
	for _, n := range networks {
		if n.Name == config.NetworkName {
			return nil
		}
	}
	_, err = cli.CreateNetwork(docker.CreateNetworkOptions{
		Name:   config.NetworkName,
		Driver: "bridge",
		Labels: map[string]string{config.LabelManaged: "true"},
	})
	if err != nil {
		return fmt.Errorf("creating network %s: %w", config.NetworkName, err)
	}
	return nil
}
