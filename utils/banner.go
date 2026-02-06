package utils

import (
	"fmt"

	"github.com/fatih/color"
)

func PrintBanner() {
	blue := color.New(color.FgHiBlue).SprintFunc()
	magenta := color.New(color.FgHiMagenta).SprintFunc()
	white := color.New(color.FgWhite).SprintFunc()
	dim := color.New(color.FgHiBlack).SprintFunc()

	// XZap ASCII art
	// X = Blue, ZAP = Purple/Magenta
	fmt.Println()
	fmt.Println(blue(" ██╗  ██╗") + magenta("███████╗ █████╗ ██████╗ "))
	fmt.Println(blue(" ╚██╗██╔╝") + magenta("╚══███╔╝██╔══██╗██╔══██╗"))
	fmt.Println(blue("  ╚███╔╝ ") + magenta("  ███╔╝ ███████║██████╔╝"))
	fmt.Println(blue("  ██╔██╗ ") + magenta(" ███╔╝  ██╔══██║██╔═══╝ "))
	fmt.Println(blue(" ██╔╝ ██╗") + magenta("███████╗██║  ██║██║     "))
	fmt.Println(blue(" ╚═╝  ╚═╝") + magenta("╚══════╝╚═╝  ╚═╝╚═╝     "))
	fmt.Println()
	fmt.Println(dim("           from ") + white("Apptitude Labs"))
	fmt.Println()
	fmt.Println(white("     The Ultimate Xcode Cleaner"))
	fmt.Println()
}
