package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"xzap/utils"
)

var cleanSimsDryRun bool
var cleanSimsForce bool

var cleanSimsCmd = &cobra.Command{
	Use:   "cleansims",
	Short: "Clean simulator devices with large data (over 2GB)",
	Run: func(cmd *cobra.Command, args []string) {
		base := utils.ExpandPath("~/Library/Developer/CoreSimulator/Devices")
		entries, err := os.ReadDir(base)
		if err != nil {
			color.Red("Failed to read simulators directory")
			return
		}

		nameMap := utils.GetSimFriendlyNames()
		var toDelete []struct {
			path string
			name string
			size int64
		}

		for _, e := range entries {
			uuid := e.Name()
			devicePath := filepath.Join(base, uuid)
			dataPath := filepath.Join(devicePath, "data")
			size := utils.CalculateDirSize(dataPath)

			if size > 2<<30 {
				var name string
				if info, found := nameMap[uuid]; found {
					name = info.Name
				} else {
					name = fmt.Sprintf("Orphaned (%s)", uuid)
				}
				toDelete = append(toDelete, struct {
					path string
					name string
					size int64
				}{devicePath, name, size})
			}
		}

		if len(toDelete) == 0 {
			color.Green("No simulators over 2GB found.")
			return
		}

		// Show what will be deleted
		fmt.Println()
		color.Cyan("Simulators over 2GB:")
		var totalSize int64
		for _, sim := range toDelete {
			color.Yellow("  • %s (%.2f GB)", sim.name, float64(sim.size)/(1<<30))
			totalSize += sim.size
		}
		fmt.Println()
		color.Cyan("Total: %d simulators, %.2f GB", len(toDelete), float64(totalSize)/(1<<30))
		fmt.Println()

		if cleanSimsDryRun {
			color.Cyan("[Dry Run] No simulators were deleted.")
			return
		}

		if !cleanSimsForce {
			reader := bufio.NewReader(os.Stdin)
			fmt.Print("Delete these simulators? [y/N]: ")
			resp, _ := reader.ReadString('\n')
			resp = strings.TrimSpace(strings.ToLower(resp))

			if resp != "y" {
				color.Yellow("Cancelled.")
				return
			}
		}

		for _, sim := range toDelete {
			color.Red("🗑️  Deleting %s...", sim.name)
			uuid := filepath.Base(sim.path)
			// Try xcrun simctl delete first to properly deregister the device
			if err := exec.Command("xcrun", "simctl", "delete", uuid).Run(); err != nil {
				// Fall back to direct removal for truly orphaned simulators
				os.RemoveAll(sim.path)
			}
		}

		color.Green("✅ Deleted %d simulators, freed %.2f GB", len(toDelete), float64(totalSize)/(1<<30))
	},
}

func init() {
	cleanSimsCmd.Flags().BoolVar(&cleanSimsDryRun, "dry-run", false, "Preview what would be deleted")
	cleanSimsCmd.Flags().BoolVar(&cleanSimsForce, "force", false, "Delete without confirmation")
	rootCmd.AddCommand(cleanSimsCmd)
}
