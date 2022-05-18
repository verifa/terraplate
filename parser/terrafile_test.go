package parser

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseTerrafile(t *testing.T) {
	tf, err := ParseTerrafile("../examples/simple/terraplate.hcl")
	require.NoError(t, err)

	t.Logf("tf: %#v", tf.ExecBlock.ExtraArgs)
	t.Logf("tf: %#v", tf.ExecBlock.PlanBlock)
}
