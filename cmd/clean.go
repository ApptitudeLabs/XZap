package cmd

import (
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"xclean/utils"
)

var dryRun bool

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Clean Xcode caches and data",
	Run: func(cmd *cobra.Command, args []string) {
		utils.PrettyTask("Cleaning DerivedData", func() {
			cleanPath("~/Library/Developer/Xcode/DerivedData")
		})
		utils.PrettyTask("Cleaning Archives", func() {
			cleanPath("~/Library/Developer/Xcode/Archives")
		})
		utils.PrettyTask("Cleaning ModuleCache", func() {
			cleanPath("~/Library/Developer/Xcode/ModuleCache.noindex")
		})
		utils.PrettyTask("Cleaning SwiftPM Cache", func() {
			cleanPath("~/Library/Caches/org.swift.swiftpm")
		})
	},
}

func init() {
	cleanCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview what would be deleted")
	rootCmd.AddCommand(cleanCmd)
}

func cleanPath(path string) {
	fullPath := utils.ExpandPath(path)
	files, err := os.ReadDir(fullPath)
	if err != nil {
		color.Yellow("⚠️  Skipping %s (not found)", path)
		return
	}
	for _, file := range files {
		fp := filepath.Join(fullPath, file.Name())
		if dryRun {
			color.Cyan("[Dry Run] Would delete: %s", fp)
		} else {
			os.RemoveAll(fp)
			color.Red("Deleted: %s", fp)
		}
	}
}
