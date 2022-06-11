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
	"github.com/verifa/terraplate/parser"
)

// parseCmd represents the parse command
var parseCmd = &cobra.Command{
	Use:   "parse",
	Short: "Parse the terraplate files and print a summary",
	Long: `Parse the terraplate files and print a summary.
	
This is useful if you want to check the configuration before running
the build command, for example.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := parser.Parse(&config.ParserConfig)
		if err != nil {
			return fmt.Errorf("parsing terraplate: %w", err)
		}
		for _, tf := range config.RootModules() {
			fmt.Println("Root Module:", tf.Path)

			data, dataErr := tf.BuildData()
			if dataErr != nil {
				return fmt.Errorf("getting build data for %s: %w", tf.Path, dataErr)
			}
			fmt.Println("Templates:")
			for _, tmpl := range tf.Templates {
				condition, condErr := tmpl.Condition(data)
				if condErr != nil {
					return fmt.Errorf("evaluating condition for template \"%s\" in %s: %w", tmpl.Name, tf.Path, condErr)
				}
				if condition {
					fmt.Printf(" - %s --> %s\n", tmpl.Name, tmpl.Target)
				}
			}
			fmt.Println("Variables:")
			for name := range tf.Variables() {
				fmt.Println(" -", name)
			}
			fmt.Println("Locals:")
			for name := range tf.Locals() {
				fmt.Println(" -", name)
			}
			fmt.Println("Values:")
			for name := range tf.Values() {
				fmt.Println(" -", name)
			}
			fmt.Println("")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(parseCmd)
}
