package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"xzap/utils"
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
		utils.PrettyTask("Cleaning Xcode Cache", func() {
			cleanPath("~/Library/Caches/com.apple.dt.Xcode")
		})
		utils.PrettyTask("Cleaning iOS DeviceSupport", func() {
			cleanPath("~/Library/Developer/Xcode/iOS DeviceSupport")
		})
		utils.PrettyTask("Cleaning DocumentationCache", func() {
			cleanPath("~/Library/Developer/Xcode/DocumentationCache")
		})
		utils.PrettyTask("Cleaning Xcode Products", func() {
			cleanPath("~/Library/Developer/Xcode/Products")
		})
		utils.PrettyTask("Cleaning SwiftPM Cache", func() {
			cleanPath("~/Library/Caches/org.swift.swiftpm")
		})
		utils.PrettyTask("Cleaning SwiftPM Data", func() {
			cleanPath("~/.swiftpm")
		})
		utils.PrettyTask("Cleaning CocoaPods Cache", func() {
			cleanPath("~/Library/Caches/CocoaPods")
		})
		utils.PrettyTask("Cleaning CocoaPods Specs", func() {
			cleanPath("~/.cocoapods/repos")
		})
		utils.PrettyTask("Cleaning Carthage Cache", func() {
			cleanPath("~/Library/Caches/org.carthage.CarthageKit")
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

	var totalSize int64
	for _, file := range files {
		fp := filepath.Join(fullPath, file.Name())
		totalSize += utils.CalculateDirSize(fp)
	}

	sizeStr := fmt.Sprintf("%.2f GB", float64(totalSize)/(1<<30))
	if totalSize < 1<<20 {
		sizeStr = fmt.Sprintf("%d KB", totalSize/1024)
	} else if totalSize < 1<<30 {
		sizeStr = fmt.Sprintf("%.2f MB", float64(totalSize)/(1<<20))
	}

	for _, file := range files {
		fp := filepath.Join(fullPath, file.Name())
		if dryRun {
			color.Cyan("[Dry Run] Would delete: %s", fp)
		} else {
			os.RemoveAll(fp)
		}
	}

	if !dryRun && totalSize > 0 {
		color.Green("  Freed %s", sizeStr)
	}
}
