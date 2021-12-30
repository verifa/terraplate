/*
Copyright © 2021 Verifa <info@verifa.io>

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

	"github.com/spf13/cobra"
	"github.com/verifa/terraplate/builder"
	"github.com/verifa/terraplate/parser"
	"github.com/verifa/terraplate/runner"
)

var skipValidate bool

// buildCmd represents the build command
var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build Terraform files based your Terrafiles",
	Long: `Build (or generate) the Terraform files.
	
For each Terrafile that is detected, build the Terraform files using the
templates and configurations detected.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := parser.Parse(&config.ParserConfig)
		if err != nil {
			return fmt.Errorf("parsing terraplate: %w", err)
		}

		if err := builder.Build(config); err != nil {
			return fmt.Errorf("building terrplate: %w", err)
		}

		if !skipValidate {
			if err := runner.Run(config, runner.RunValidate()); err != nil {
				return fmt.Errorf("validation failed: %w", err)
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)

	buildCmd.Flags().BoolVar(&skipValidate, "skip-validate", false, "Skip calling terraform validate after build")
}
