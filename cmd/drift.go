/*
Copyright © 2022 Verifa <info@verifa.io>

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
	"github.com/verifa/terraplate/notify"
	"github.com/verifa/terraplate/parser"
	"github.com/verifa/terraplate/runner"
)

var (
	notifyResultsUrl  string
	notifyType        string
	notifyFilter      string
	notifySlackConfig = notify.DefaultSlackConfig()

	repo notify.Repo
)

var driftCmd = &cobra.Command{
	Use:   "drift",
	Short: "Detect drift in your infrastructure (experimental feature)",
	Long:  `Detect drift in your infrastructure and send Slack notifications`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var notifyService notify.Service
		if notifyType != "" {
			var notifyErr error
			notifyService, notifyErr = notify.New(
				notify.WithNotify(notify.NotifyType(notifyType)),
				notify.NotifyOn(notify.NotifyFilter(notifyFilter)),
				notify.WithSlackConfig(notifySlackConfig),
			)
			if notifyErr != nil {
				return fmt.Errorf("creating notification service: %w", notifyErr)
			}

		}

		// Parse
		config, err := parser.Parse(&config.ParserConfig)
		if err != nil {
			return fmt.Errorf("parsing terraplate: %w", err)
		}
		// Plan
		fmt.Print(terraformStartMessage)
		runOpts := []func(r *runner.TerraRunOpts){
			runner.RunBuild(),
			runner.RunInit(),
			runner.RunPlan(),
			runner.RunShowPlan(),
			runner.Jobs(planJobs),
		}
		runOpts = append(runOpts, runner.ExtraArgs(args))
		r := runner.Run(config, runOpts...)

		if notifyService != nil {
			repo, repoErr := notify.LookupRepo(
				notify.WithRepo(repo),
			)
			if repoErr != nil {
				return fmt.Errorf("looking up repository details: %w", repoErr)
			}
			sendErr := notifyService.Send(&notify.Data{
				Runner:     r,
				Repo:       repo,
				ResultsURL: notifyResultsUrl,
			})
			if sendErr != nil {
				return fmt.Errorf("sending notification: %w", sendErr)
			}
		}

		fmt.Print(r.Summary(runner.OutputLevelAll))
		return nil
	},
}

func init() {
	RootCmd.AddCommand(driftCmd)

	driftCmd.Flags().IntVarP(&planJobs, "jobs", "j", runner.DefaultJobs, "Number of concurrent terraform jobs to run at one time")
	driftCmd.Flags().StringVar(&notifyType, "notify", "", "Notification type (only slack supported)")
	driftCmd.Flags().StringVar(&notifyFilter, "notify-on", "", "When to send a notification (possible values are \"all\" and \"drift\")")
	driftCmd.Flags().StringVar(&notifySlackConfig.Channel, "slack-channel", "", "Slack channel where to send the notification (required if notify=slack)")
	driftCmd.Flags().StringVar(&notifyResultsUrl, "results-url", "", "Provide a custom URL that will be shown in the notification (such as a link to your CI log for easy access)")
	driftCmd.Flags().StringVar(&repo.Name, "repo-name", "", "Name of the repository to show in notifications")
	driftCmd.Flags().StringVar(&repo.Branch, "repo-branch", "", "Branch of the repository to show in notifications")
}
