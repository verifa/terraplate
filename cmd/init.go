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

var (
	initJobs      int
	initSkipBuild bool
	initUpgrade   bool
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Runs terraform init on all subdirectories",
	Long:  `Runs terraform init on all subdirectories.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := parser.Parse(&config.ParserConfig)
		if err != nil {
			return fmt.Errorf("parsing terraplate: %w", err)
		}
		runOpts := []func(r *runner.TerraRunOpts){
			runner.RunInit(),
			runner.Jobs(initJobs),
			runner.ExtraArgs(args),
		}

		if !initSkipBuild {
			runOpts = append(runOpts, runner.RunBuild())
		}
		if initUpgrade {
			runOpts = append(runOpts, runner.RunInitUpgrade())
		}

		runOpts = append(runOpts, runner.ExtraArgs(args))
		r := runner.Run(config, runOpts...)
		fmt.Println(r.Log(runner.OutputLevelAll))
		fmt.Println(r.Summary(runner.OutputLevelAll))

		return r.Errors()
	},
}

func init() {
	RootCmd.AddCommand(initCmd)

	initCmd.Flags().BoolVar(&initSkipBuild, "skip-build", false, "Skip build process (default: false)")
	initCmd.Flags().BoolVarP(&initUpgrade, "upgrade", "u", false, "Perform upgrade when initializing")
	initCmd.Flags().IntVarP(&initJobs, "jobs", "j", runner.DefaultJobs, "Number of concurrent terraform jobs to run at one time")
}
