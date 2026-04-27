package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/spf13/cobra"

	"github.com/clawfleet/clawfleet/internal/config"
	"github.com/clawfleet/clawfleet/internal/container"
	"github.com/clawfleet/clawfleet/internal/port"
	"github.com/clawfleet/clawfleet/internal/snapshot"
	"github.com/clawfleet/clawfleet/internal/state"
)

var (
	pullFlag         bool
	fromSnapshotFlag string
	runtimeFlag      string
)

var createCmd = &cobra.Command{
	Use:     "create <N>",
	Short:   "Create N isolated instances (OpenClaw or Hermes)",
	Args:    cobra.ExactArgs(1),
	Example: "  clawfleet create 3\n  clawfleet create 1 --runtime hermes\n  clawfleet create 3 --pull",
	RunE:    runCreate,
}

func init() {
	createCmd.Flags().BoolVar(&pullFlag, "pull", false, "Force re-pull image from registry (even if already present locally)")
	createCmd.Flags().StringVar(&fromSnapshotFlag, "from-snapshot", "", "Create instance from a saved snapshot (OpenClaw only)")
	createCmd.Flags().StringVar(&runtimeFlag, "runtime", "openclaw", "Runtime to create: openclaw or hermes")
}

func runCreate(cmd *cobra.Command, args []string) error {
	n, err := strconv.Atoi(args[0])
	if err != nil || n < 1 {
		return fmt.Errorf("N must be a positive integer")
	}

	if runtimeFlag != "openclaw" && runtimeFlag != "hermes" {
		return fmt.Errorf("--runtime must be 'openclaw' or 'hermes' (got %q)", runtimeFlag)
	}

	if fromSnapshotFlag != "" && runtimeFlag == "hermes" {
		return fmt.Errorf("--from-snapshot is not supported for Hermes (Soul Archive is OpenClaw-only)")
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	cli, err := container.NewClient()
	if err != nil {
		return err
	}

	// Pick runtime-specific image and pull args.
	var imageRef, imageName, imageTag string
	if runtimeFlag == "hermes" {
		imageRef = cfg.HermesImageRef()
		imageName = cfg.Hermes.ImageName
		imageTag = cfg.Hermes.ImageTag
	} else {
		imageRef = cfg.ImageRef()
		imageName = cfg.Image.Name
		imageTag = cfg.Image.Tag
	}

	// Check image exists; auto-pull if missing, force-pull if --pull flag set.
	exists, err := container.ImageExists(cli, imageRef)
	if err != nil {
		return err
	}
	if !exists || pullFlag {
		if !exists {
			fmt.Printf("Image %s not found locally, pulling from registry...\n", imageRef)
		} else {
			fmt.Printf("Pulling latest %s from registry...\n", imageRef)
		}
		if pullErr := container.PullImage(cli, imageName, imageTag, os.Stdout); pullErr != nil {
			return fmt.Errorf("pull failed: %v\nRun 'clawfleet build' to build it manually", pullErr)
		}
		fmt.Println("Image pulled successfully.")
	}

	// Ensure network
	if err := container.EnsureNetwork(cli); err != nil {
		return err
	}

	// Load state
	store, err := state.Load()
	if err != nil {
		return err
	}

	// Parse resource limits
	memBytes, err := container.ParseMemoryBytes(cfg.Resources.MemoryLimit)
	if err != nil {
		return err
	}
	nanoCPUs := int64(cfg.Resources.CPULimit * 1e9)

	dataDir, err := config.DataDir()
	if err != nil {
		return err
	}

	dataSuffix := "openclaw"
	if runtimeFlag == "hermes" {
		dataSuffix = "hermes"
	}

	created := 0
	firstName := ""
	for i := 0; i < n; i++ {
		name := store.NextName(config.NamingPrefix(runtimeFlag))
		if firstName == "" {
			firstName = name
		}
		usedPorts := store.UsedPorts()

		novncPort, err := port.FindAvailable(cfg.Ports.NoVNCBase, usedPorts)
		if err != nil {
			return fmt.Errorf("allocating noVNC port: %w", err)
		}
		usedPorts[novncPort] = true

		gatewayPort, err := port.FindAvailable(cfg.Ports.GatewayBase, usedPorts)
		if err != nil {
			return fmt.Errorf("allocating gateway port: %w", err)
		}

		instanceDataDir := filepath.Join(dataDir, "data", name, dataSuffix)
		if err := os.MkdirAll(instanceDataDir, 0755); err != nil {
			return fmt.Errorf("creating data dir for %s: %w", name, err)
		}

		// Load snapshot data if specified (OpenClaw only — guarded above)
		if fromSnapshotFlag != "" {
			if err := snapshot.Load(fromSnapshotFlag, instanceDataDir); err != nil {
				return fmt.Errorf("loading snapshot: %w", err)
			}
		}

		fmt.Printf("Creating %s ... ", name)

		containerID, err := container.Create(cli, container.CreateParams{
			Name:        name,
			ImageRef:    imageRef,
			NoVNCPort:   novncPort,
			GatewayPort: gatewayPort,
			DataDir:     instanceDataDir,
			MemoryBytes: memBytes,
			NanoCPUs:    nanoCPUs,
			RuntimeType: runtimeFlag,
		})
		if err != nil {
			fmt.Println("✗")
			return err
		}

		if err := container.Start(cli, containerID); err != nil {
			fmt.Println("✗")
			return fmt.Errorf("starting %s: %w", name, err)
		}

		inst := &state.Instance{
			Name:        name,
			ContainerID: containerID,
			Status:      "running",
			Ports:       state.Ports{NoVNC: novncPort, Gateway: gatewayPort},
			CreatedAt:   time.Now(),
			RuntimeType: runtimeFlag,
		}
		store.Add(inst)
		if err := store.Save(); err != nil {
			return fmt.Errorf("saving state: %w", err)
		}

		// Associate model asset from snapshot if available
		if fromSnapshotFlag != "" {
			if snapStore, err := state.LoadSnapshots(); err == nil {
				if snapMeta := snapStore.GetByName(fromSnapshotFlag); snapMeta != nil && snapMeta.ModelAssetID != "" {
					store.SetConfig(name, snapMeta.ModelAssetID, "", "")
					_ = store.Save()
				}
			}
		}

		if runtimeFlag == "hermes" {
			fmt.Printf("✓  dashboard: http://localhost:%d\n", novncPort)
		} else {
			fmt.Printf("✓  desktop: http://localhost:%d\n", novncPort)
		}
		created++
	}

	if runtimeFlag == "hermes" {
		fmt.Printf("\n%d Hermes instance(s) ready. Run 'clawfleet shell %s' to start chatting.\n",
			created, firstName)
	} else {
		fmt.Printf("\n%d claw(s) ready. Run 'clawfleet desktop %s' to open the desktop.\n",
			created, firstName)
	}
	return nil
}
