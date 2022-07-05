package runner

import "github.com/fatih/color"

var (
	boldColor          = color.New(color.Bold)
	errorColor         = color.New(color.FgRed, color.Bold)
	runCancelled       = color.New(color.FgRed, color.Bold)
	planNotAvailable   = color.New(color.FgMagenta, color.Bold)
	planNoChangesColor = color.New(color.FgGreen, color.Bold)
	planCreateColor    = color.New(color.FgGreen, color.Bold)
	planDestroyColor   = color.New(color.FgRed, color.Bold)
	planUpdateColor    = color.New(color.FgYellow, color.Bold)
)

var (
	textSeparator = boldColor.Sprint("\n─────────────────────────────────────────────────────────────────────────────\n\n")
)
