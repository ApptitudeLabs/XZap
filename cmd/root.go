package cmd

import (
	"fmt"
	"os"
	"xclean/utils"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "xclean",
	Short: "🧹 xclean by Apptitude Labs — Clean up Xcode junk fast!",
	Long: `
🧹 xclean by Apptitude Labs

The fastest way to clean your Xcode workspace, simulators, and runtimes.
`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		utils.PrintBanner()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
