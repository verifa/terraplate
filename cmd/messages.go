package cmd

import (
	"github.com/fatih/color"
)

var (
	errorColor            = color.New(color.FgRed, color.Bold)
	boldText              = color.New(color.Bold)
	terraformStartMessage = boldText.Sprint("\nTerraforming root modules...\n\n")
	devStartMessage       = boldText.Sprint("\nStarting dev mode...\n\n")
)
