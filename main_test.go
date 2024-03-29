package main

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/verifa/terraplate/parser"
	"github.com/verifa/terraplate/runner"
)

func TestMain(t *testing.T) {
	type test struct {
		dir           string
		skipTerraform bool
	}
	tests := []test{
		{
			dir: "examples/simple",
		},
		{
			dir: "examples/simple/dev",
		},
		{
			dir: "examples/complete",
		},
		{
			dir: "examples/aws",
		},
		{
			dir: "examples/aws/stack",
		},
		{
			dir: "examples/config",
		},
		{
			dir: "examples/nested",
		},
	}
	for _, tc := range tests {
		t.Run(tc.dir, func(t *testing.T) {
			config, err := parser.Parse(&parser.Config{
				Chdir: tc.dir,
			})
			require.NoError(t, err)

			if !tc.skipTerraform {
				runner := runner.Run(config,
					runner.RunBuild(),
					runner.RunValidate(),
					runner.RunInit(),
					runner.RunPlan(),
					runner.RunApply())

				require.NoError(t, runner.Errors())
			}
		})
	}
}
