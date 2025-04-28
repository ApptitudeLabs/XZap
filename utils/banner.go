package utils

import (
	"fmt"

	"github.com/common-nighthawk/go-figure"
	"github.com/fatih/color"
)

func PrintBanner() {
	banner := figure.NewFigure("xclean", "slant", true)
	white := color.New(color.FgWhite).SprintFunc()

	fmt.Println()
	color.Set(color.FgCyan)
	banner.Print()
	color.Unset()

	fmt.Println(white("        by Apptitude Labs"))
	fmt.Println()
	fmt.Println(white("    The fastest way to clean your Xcode workspace, simulators, and runtimes."))
	fmt.Println()
}
