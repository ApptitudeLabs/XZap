package utils

import (
	"fmt"
	"os"
	"os/exec"
)

// CheckXcrun checks if xcrun is available and exits with a helpful message if not
func CheckXcrun() {
	if _, err := exec.LookPath("xcrun"); err != nil {
		fmt.Println()
		fmt.Println("Error: Xcode Command Line Tools are required")
		fmt.Println()
		fmt.Println("xcrun was not found on your system.")
		fmt.Println("Please install Xcode Command Line Tools by running:")
		fmt.Println()
		fmt.Println("  xcode-select --install")
		fmt.Println()
		os.Exit(1)
	}
}

// IsXcrunAvailable returns true if xcrun is available
func IsXcrunAvailable() bool {
	_, err := exec.LookPath("xcrun")
	return err == nil
}
