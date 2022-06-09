package cmd

import (
	"fmt"

	"github.com/fatih/color"
)

var (
	successColor          = color.New(color.FgGreen, color.Bold)
	errorColor            = color.New(color.FgRed, color.Bold)
	boldText              = color.New(color.Bold)
	buildStartMessage     = boldText.Sprint("\nBuilding root modules...\n\n")
	buildSuccessMessage   = fmt.Sprintf("\n%s All root modules built\n\n", successColor.Sprint("Success!"))
	terraformStartMessage = boldText.Sprint("\nTerraforming root modules...\n\n")
)
