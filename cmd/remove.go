package cmd

import (
	"strings"
	"xzap/utils"

	"github.com/spf13/cobra"
)

var removeRuntimeCmd = &cobra.Command{
	Use:   "runtime [name]",
	Short: "Remove a specific Xcode runtime",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := strings.Join(args, " ")
		utils.RemoveRuntime(name, forceClean)
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)
	removeCmd.AddCommand(removeRuntimeCmd)
	removeRuntimeCmd.Flags().BoolVar(&forceClean, "force", false, "Force deletion without confirmation")
}

var removeCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove simulators or runtimes",
}
