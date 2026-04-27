package cli

import "github.com/spf13/cobra"

var snapshotCmd = &cobra.Command{
	Use:   "snapshot",
	Short: "Manage OpenClaw instance snapshots",
	Long: `Save, list, and delete snapshots of OpenClaw instances.
Snapshots (Soul Archive) are an OpenClaw-only feature; Hermes instances are not supported.`,
}

func init() {
	snapshotCmd.AddCommand(snapshotSaveCmd, snapshotListCmd, snapshotDeleteCmd)
}
