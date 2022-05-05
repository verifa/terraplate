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
	"strings"

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

			fmt.Println("Templates:")
			for _, tmpl := range tf.BuildTemplates() {
				fmt.Printf(" - %s: %s --> %s\n", tmpl.Name, strings.Join(tmpl.SourceFiles(), ","), tmpl.BuildTarget())
			}
			fmt.Println("Variables:")
			for name := range tf.BuildVariables() {
				fmt.Println(" -", name)
			}
			fmt.Println("Locals:")
			for name := range tf.BuildLocals() {
				fmt.Println(" -", name)
			}
			buildValues, err := tf.BuildValues()
			if err != nil {
				return fmt.Errorf("getting build values for %s: %w", tf.Path, err)
			}
			fmt.Println("Values:")
			for name := range buildValues {
				fmt.Println(" -", name)
			}
			fmt.Println("")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(parseCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// parseCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// parseCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
