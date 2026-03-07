package cli

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/weiyong1024/clawsandbox/internal/state"
)

var desktopCmd = &cobra.Command{
	Use:   "desktop <name>",
	Short: "Open a lobster's noVNC desktop in the browser",
	Args:  cobra.ExactArgs(1),
	RunE:  runDesktop,
}

func runDesktop(cmd *cobra.Command, args []string) error {
	store, err := state.Load()
	if err != nil {
		return err
	}

	inst := store.Get(args[0])
	if inst == nil {
		return fmt.Errorf("instance not found: %s", args[0])
	}
	if inst.Status != "running" {
		return fmt.Errorf("%s is not running (status: %s)", inst.Name, inst.Status)
	}

	url := fmt.Sprintf("http://localhost:%d", inst.Ports.NoVNC)
	fmt.Printf("Opening %s desktop at %s\n", inst.Name, url)
	return openBrowser(url)
}

func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	default:
		fmt.Printf("Open this URL in your browser: %s\n", url)
		return nil
	}
	return cmd.Start()
}
