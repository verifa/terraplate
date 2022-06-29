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

// devCmd represents the plan command
var devCmd = &cobra.Command{
	Use:   "dev",
	Short: "Enters dev mode which launches a Terminal UI for Terraplate",
	Long:  `Enters dev mode which launches a Terminal UI for building and running Terraplate root modules.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Parse
		config, err := parser.Parse(&config.ParserConfig)
		if err != nil {
			return fmt.Errorf("parsing terraplate: %w", err)
		}

		// Dev mode
		fmt.Print(devStartMessage)
		runOpts := []func(r *runner.TerraRunOpts){
			runner.RunBuild(),
			runner.RunInit(),
			runner.RunPlan(),
			runner.RunShowPlan(),
			runner.Jobs(planJobs),
			// Discard any output from the runner itself.
			// This does not discard the Terraform output.
			runner.Output(io.Discard),
		}
		if runInit {
			runOpts = append(runOpts, runner.RunInit())
		}
		runOpts = append(runOpts, runner.ExtraArgs(args))
		runner := runner.New(config, runOpts...)

		p := tea.NewProgram(
			tui.New(runner),
			tea.WithAltScreen(),
			tea.WithMouseCellMotion(),
		)
		if err := p.Start(); err != nil {
			fmt.Printf("Alas, there's been an error: %v", err)
			os.Exit(1)
		}
		return nil
	},
}

func init() {
	RootCmd.AddCommand(devCmd)

	devCmd.Flags().IntVarP(&planJobs, "jobs", "j", runner.DefaultJobs, "Number of concurrent terraform jobs to run at one time")
}
