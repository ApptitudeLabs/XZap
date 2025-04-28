package cmd

import (
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"xclean/utils"
)

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
		for _, e := range entries {
			devicePath := filepath.Join(base, e.Name())
			dataPath := filepath.Join(devicePath, "data")
			size := utils.CalculateDirSize(dataPath)
			if size > 2<<30 {
				color.Yellow("🗑️ Deleting simulator %s (%.2f GB)", e.Name(), float64(size)/(1<<30))
				os.RemoveAll(devicePath)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(cleanSimsCmd)
}
