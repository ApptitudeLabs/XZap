package cmd

import (
	"fmt"
	"os"
	"xzap/utils"

	"github.com/spf13/cobra"
)

// Version is set at build time via ldflags
var Version = "dev"

var rootCmd = &cobra.Command{
	Use:     "xzap",
	Short:   "⚡ XZap by Apptitude Labs — The Ultimate Xcode Cleaner",
	Version: Version,
	Long: `
⚡ XZap by Apptitude Labs — The Ultimate Xcode Cleaner

The fastest way to clean your Xcode workspace, simulators, and runtimes.
`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		utils.PrintBanner()
		utils.CheckXcrun()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
