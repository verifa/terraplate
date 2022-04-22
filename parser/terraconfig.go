package parser

import (
	"fmt"

	"github.com/imdario/mergo"
)

type TerraConfig struct {
	Terrafiles []*Terrafile
}

// RootModules returns the Terrafiles that are considered root modules
// and should therefore be processed
func (c *TerraConfig) RootModules() []*Terrafile {
	var files = make([]*Terrafile, 0)
	for _, tf := range c.Terrafiles {
		if tf.IsRoot {
			files = append(files, tf)
		}
	}
	return files
}

// RootTerrafile returns the top-most (root) Terrafiles.
// It's possible that there's multiple root terrafiles if there are multiple
// trees of terrafiles that have been parsed
func (c *TerraConfig) RootTerrafiles() []*Terrafile {
	if len(c.Terrafiles) == 0 {
		// Not possible as an empty list of terrafiles should through an error
		// during parsing
		return nil
	}
	// Create a map of unique root terrafiles
	var tfMap = make(map[string]*Terrafile)
	for _, tf := range c.Terrafiles {
		rootTf := tf.rootAncestor()
		// Doesn't matter if we overwrite as it's the same root terrafile
		tfMap[rootTf.Path] = rootTf
	}
	var rootTfs = make([]*Terrafile, 0, len(tfMap))
	for _, tf := range tfMap {
		rootTfs = append(rootTfs, tf)
	}
	return rootTfs
}

func (c *TerraConfig) MergeTerrafiles() error {
	// Logic is to get the topmost (root) Terrafile and traverse down the tree
	// merging parent with child terrafiles as we go
	rootTfs := c.RootTerrafiles()

	for _, rootTf := range rootTfs {
		// Set defaults for root terrafile
		if err := mergo.Merge(rootTf, defaultTerrafile); err != nil {
			return fmt.Errorf("setting defaults for root terrafile %s: %w", rootTf.Path, err)
		}

		travErr := rootTf.traverseChildren(func(parent *Terrafile, tf *Terrafile) error {
			if err := tf.mergeTerrafile(parent); err != nil {
				return fmt.Errorf("merging terrafile %s with parent %s: %w", tf.Path, parent.Path, err)
			}
			return nil
		})
		if travErr != nil {
			return fmt.Errorf("traversing terrafiles from root %s: %w", rootTf.Path, travErr)
		}
	}
	return nil
}
