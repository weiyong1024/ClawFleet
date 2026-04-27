package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "clawfleet",
	Short: "Deploy and manage a fleet of OpenClaw and Hermes instances",
	Long: `ClawFleet lets you spin up isolated OpenClaw and Hermes agent instances
on a single machine. Each instance runs in its own Docker container,
managed from a unified browser dashboard or this CLI.`,
}

func Execute() {
	rootCmd.AddCommand(
		buildCmd,
		createCmd,
		listCmd,
		startCmd,
		stopCmd,
		restartCmd,
		destroyCmd,
		desktopCmd,
		logsCmd,
		dashboardCmd,
		configCmd,
		configureCmd,
		shellCmd,
		snapshotCmd,
		versionCmd,
	)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
