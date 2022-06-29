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
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/verifa/terraplate/parser"
	"github.com/verifa/terraplate/runner"
	"github.com/verifa/terraplate/tui"
)

var (
	planSkipBuild bool
	runInit       bool
	planJobs      int
	planDevMode   bool
)

// planCmd represents the plan command
var planCmd = &cobra.Command{
	Use:   "plan",
	Short: "Runs terraform plan on all subdirectories",
	Long:  `Runs terraform plan on all subdirectories.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := parser.Parse(&config.ParserConfig)
		if err != nil {
			return fmt.Errorf("parsing terraplate: %w", err)
		}
		fmt.Print(terraformStartMessage)
		runOpts := []func(r *runner.TerraRunOpts){
			runner.RunPlan(),
			runner.RunShowPlan(),
			runner.Jobs(planJobs),
		}
		if !planSkipBuild {
			runOpts = append(runOpts, runner.RunBuild())
		}
		if runInit {
			runOpts = append(runOpts, runner.RunInit())
		}
		runOpts = append(runOpts, runner.ExtraArgs(args))
		r := runner.Run(config, runOpts...)

		if planDevMode {
			// Start dev mode
			fmt.Print(devStartMessage)
			runOpts := []func(r *runner.TerraRunOpts){
				runner.RunBuild(),
				runner.RunPlan(),
				runner.RunShowPlan(),
				runner.Jobs(planJobs),
				// Discard any output from the runner itself.
				// This does not discard the Terraform output.
				runner.Output(io.Discard),
			}
			// Override options in runner
			r.Opts = runner.NewOpts(runOpts...)
			p := tea.NewProgram(
				tui.New(r),
				tea.WithAltScreen(),
				tea.WithMouseCellMotion(),
			)
			if err := p.Start(); err != nil {
				fmt.Printf("Alas, there's been an error: %v", err)
				os.Exit(1)
			}

			return nil
		}

		// Print log
		fmt.Println(r.Log())

		fmt.Println(r.Summary())

		return r.Errors()
	},
}

func init() {
	RootCmd.AddCommand(planCmd)

	planCmd.Flags().BoolVar(&planSkipBuild, "skip-build", false, "Skip build process (default: false)")
	planCmd.Flags().BoolVar(&runInit, "init", false, "Run terraform init also")
	planCmd.Flags().BoolVar(&planDevMode, "dev", false, "Start dev mode after plan finishes")
	planCmd.Flags().IntVarP(&planJobs, "jobs", "j", runner.DefaultJobs, "Number of concurrent terraform jobs to run at one time")
}
