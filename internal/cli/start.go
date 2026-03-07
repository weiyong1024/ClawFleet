package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/weiyong1024/clawsandbox/internal/container"
	"github.com/weiyong1024/clawsandbox/internal/state"
)

var startCmd = &cobra.Command{
	Use:   "start <name|all>",
	Short: "Start a stopped lobster instance",
	Args:  cobra.ExactArgs(1),
	RunE:  runStart,
}

func runStart(cmd *cobra.Command, args []string) error {
	store, err := state.Load()
	if err != nil {
		return err
	}

	cli, err := container.NewClient()
	if err != nil {
		return err
	}

	targets := resolveTargets(store, args[0])
	if len(targets) == 0 {
		return fmt.Errorf("no instance found: %s", args[0])
	}

	for _, inst := range targets {
		fmt.Printf("Starting %s ... ", inst.Name)
		if err := container.Start(cli, inst.ContainerID); err != nil {
			fmt.Println("✗")
			return err
		}
		inst.Status = "running"
		fmt.Println("✓")
	}

	return store.Save()
}
