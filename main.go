package main

import (
	"os"

	"xclean/cmd"
	"xclean/tui"
)

func main() {
	// Check for --cli flag to use legacy CLI mode
	if len(os.Args) > 1 && os.Args[1] == "--cli" {
		// Strip --cli and run legacy CLI
		os.Args = append(os.Args[:1], os.Args[2:]...)
		cmd.Execute()
		return
	}

	// Default: Launch TUI
	tui.Run()
}
