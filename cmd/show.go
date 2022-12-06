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
	"io"

	"github.com/spf13/cobra"
	"github.com/verifa/terraplate/parser"
	"github.com/verifa/terraplate/runner"
)

var (
	showCmdJobs         int
	showCmdOutputLevel  string
	showCmdShowProgress bool
)

// showCmd represents the show command
var showCmd = &cobra.Command{
	Use:   "show",
	Short: "Runs terraform show on all subdirectories",
	Long:  `Runs terraform show on all subdirectories.`,
	RunE: func(cmd *cobra.Command, args []string) error {

		outputLevel, err := runner.OutputLevel(showCmdOutputLevel).Validate()
		if err != nil {
			return err
		}

		config, err := parser.Parse(&config.ParserConfig)
		if err != nil {
			return fmt.Errorf("parsing terraplate: %w", err)
		}
		if showCmdShowProgress {
			fmt.Print(terraformStartMessage)
		}
		runOpts := []func(r *runner.TerraRunOpts){
			runner.RunShow(),
			runner.RunShowPlan(),
			runner.Jobs(showCmdJobs),
		}
		if !showCmdShowProgress {
			runOpts = append(runOpts, runner.Output(io.Discard))
		}

		runOpts = append(runOpts, runner.ExtraArgs(args))
		r := runner.Run(config, runOpts...)

		fmt.Println(r.Log(outputLevel))

		fmt.Println(r.Summary(outputLevel))

		return r.Errors()
	},
}

func init() {
	RootCmd.AddCommand(showCmd)

	showCmd.Flags().IntVarP(&showCmdJobs, "jobs", "j", runner.DefaultJobs, "Number of concurrent terraform jobs to run at one time")
	showCmd.Flags().StringVar(&showCmdOutputLevel, "output-level", string(runner.OutputLevelAll), "Level of output to show (all or drift)")
	showCmd.Flags().BoolVar(&showCmdShowProgress, "show-progress", true, "Whether to show Terraform run progress")
}
