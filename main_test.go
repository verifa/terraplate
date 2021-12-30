package main

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/verifa/terraplate/builder"
	"github.com/verifa/terraplate/parser"
	"github.com/verifa/terraplate/runner"
)

func TestMain(t *testing.T) {
	config, err := parser.Parse(&parser.Config{
		Chdir: "examples/simple",
	})
	require.NoError(t, err)
	buildErr := builder.Build(config)
	require.NoError(t, buildErr)

	runErr := runner.Run(config,
		runner.RunValidate(),
		runner.RunInit(),
		runner.RunPlan(),
		runner.RunApply())
	require.NoError(t, runErr)
}
