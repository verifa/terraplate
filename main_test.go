package main

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/verifa/terraplate/builder"
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
			dir:           "examples/aws",
			skipTerraform: true, // would require aws auth
		},
		{
			dir:           "examples/aws/stack",
			skipTerraform: true, // would require aws auth
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
			buildErr := builder.Build(config)
			require.NoError(t, buildErr)

			if !tc.skipTerraform {
				runErr := runner.Run(config,
					runner.RunValidate(),
					runner.RunInit(),
					runner.RunPlan(),
					runner.RunApply())
				require.NoError(t, runErr)
			}
		})
	}
}
