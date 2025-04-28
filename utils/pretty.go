package utils

import (
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
)

func PrettyTask(name string, task func()) {
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = "  " + name
	s.Color("cyan")
	s.Start()

	task()

	s.Stop()
	color.Green("✅  %s done!", name)
}