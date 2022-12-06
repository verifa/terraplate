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
	"github.com/verifa/terraplate/parser"
	"github.com/verifa/terraplate/runner"
)

var doValidate bool

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

		runOpts := []func(r *runner.TerraRunOpts){
			runner.RunBuild(),
		}
		if doValidate {
			runOpts = append(runOpts, runner.RunValidate())
		}
		runOpts = append(runOpts, runner.ExtraArgs(args))
		r := runner.Run(config, runOpts...)

		fmt.Println(r.Log(runner.OutputLevelAll))
		fmt.Println(r.Summary(runner.OutputLevelAll))

		return r.Errors()
	},
}

func init() {
	RootCmd.AddCommand(buildCmd)
	buildCmd.Flags().BoolVar(&doValidate, "validate", false, "Validate (requires init) each root module after build")
}
