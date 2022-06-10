/*
Copyright © 2021 Verifa <support@verifa.io>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/verifa/terraplate/parser"
)

type cmdConfig struct {
	ParserConfig parser.Config
}

var config cmdConfig

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "terraplate",
	Short: "DRY Terraform using Go Templates",
	Long: `DRY Terraform using Go Templates.

Terraplate keeps your Terraform DRY.
Create templates that get built using Go Templates to avoid repeating common
Terraform configurations like providers and backend.`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the RootCmd.
func Execute() {
	err := RootCmd.Execute()
	if err != nil {
		// Pretty print the error before finishing
		fmt.Printf("\n%s %s\n", errorColor.Sprint("Error:"), err.Error())
		os.Exit(1)
	}
}

func init() {
	RootCmd.PersistentFlags().StringVarP(&config.ParserConfig.Chdir, "chdir", "C", ".", "Switch to a different working directory before executing the given subcommand.")
}
