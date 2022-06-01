package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParser(t *testing.T) {
	// Simply test parsing all the examples...
	_, err := Parse(&Config{
		Chdir: "../examples/nested",
	})
	require.NoError(t, err)
}

func TestTraverse(t *testing.T) {
	config, err := Parse(&Config{
		Chdir: "testdata/nested/1/2",
	})
	require.NoError(t, err)
	// There should be only one root module
	require.Len(t, config.RootModules(), 1)

	rootMod := config.RootModules()[0]

	var (
		upPath   = make([]*Terrafile, 0)
		downPath = make([]*Terrafile, 0)
	)

	rootMod.traverseAncestors(func(ancestor *Terrafile) error {
		upPath = append(upPath, ancestor)
		return nil
	})
	rootMod.traverseAncestorsReverse(func(ancestor *Terrafile) error {
		// Reverse the down path, which should equal the up path
		downPath = append([]*Terrafile{ancestor}, downPath...)
		return nil
	})

	assert.Equal(t, upPath, downPath)
}

func TestOverride(t *testing.T) {
	_, err := Parse(&Config{
		Chdir: "testdata/override",
	})
	require.NoError(t, err)
}

func TestTemplates(t *testing.T) {
	_, err := Parse(&Config{
		Chdir: "testdata/testTemplates",
	})
	require.NoError(t, err)
}
