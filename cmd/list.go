package cmd

import (
	"fmt"
	"os"
	"xclean/utils"

	"github.com/spf13/cobra"
)

var showAll bool
var showCriticalOnly bool
var thresholdGB int
var outputFile string
var cleanInteractive bool
var listDryRun bool
var forceClean bool
var summaryOnly bool

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List Xcode simulator and runtime data",
}

var listSimsCmd = &cobra.Command{
	Use:   "sims",
	Short: "List simulators with their disk usage",
	Run: func(cmd *cobra.Command, args []string) {
		if showCriticalOnly && thresholdGB > 0 {
			fmt.Println("⚠️  Cannot use --critical and --threshold together. Please choose one.")
			os.Exit(1)
		}
		utils.ListSimulators(showAll, showCriticalOnly, thresholdGB, outputFile, cleanInteractive, listDryRun, forceClean, summaryOnly)
	},
}

var listRuntimesCmd = &cobra.Command{
	Use:   "runtimes",
	Short: "List installed runtimes and their sizes",
	Run: func(cmd *cobra.Command, args []string) {
		utils.ListRuntimes()
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.AddCommand(listSimsCmd)
	listCmd.AddCommand(listRuntimesCmd)

	listSimsCmd.Flags().BoolVar(&showAll, "all", false, "Show all simulators, even 0.00 GB ones")
	listSimsCmd.Flags().BoolVar(&showCriticalOnly, "critical", false, "Only show simulators larger than 3 GB")
	listSimsCmd.Flags().IntVar(&thresholdGB, "threshold", 0, "Only show simulators larger than X GB")
	listSimsCmd.Flags().StringVar(&outputFile, "output", "", "Save output to a file")
	listSimsCmd.Flags().BoolVar(&cleanInteractive, "clean", false, "Interactively clean large simulators")
	listSimsCmd.Flags().BoolVar(&listDryRun, "dry-run", false, "Simulate cleaning without deleting anything")
	listSimsCmd.Flags().BoolVar(&forceClean, "force-clean", false, "Force delete without confirmation")
	listSimsCmd.Flags().BoolVar(&summaryOnly, "summary-only", false, "Only show summary (no device list)")
}