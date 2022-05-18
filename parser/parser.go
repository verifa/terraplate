package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

func DefaultConfig() *Config {
	return &Config{
		Chdir: ".",
	}
}

type Config struct {
	Chdir string
}

func Parse(config *Config) (*TerraConfig, error) {
	ancestor, travErr := walkUpDirectory(config.Chdir)
	if travErr != nil {
		return nil, fmt.Errorf("looking for parent terraplate.hcl files: %w", travErr)
	}

	terrafiles, walkErr := walkDownDirectory(config.Chdir, ancestor)
	if walkErr != nil {
		return nil, fmt.Errorf("looking for terraplate.hcl files: %w", walkErr)
	}

	return &TerraConfig{
		Terrafiles: terrafiles,
	}, nil
}

func walkDownDirectory(dir string, ancestor *Terrafile) ([]*Terrafile, error) {
	var (
		terrafiles []*Terrafile
		terrafile  *Terrafile
		subDirs    []string
	)

	// Skip the .terraform directories
	if filepath.Base(dir) == ".terraform" {
		return terrafiles, nil
	}
	entries, readErr := os.ReadDir(dir)
	if readErr != nil {
		return nil, fmt.Errorf("reading directory \"%s\": %w", dir, readErr)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			subDirs = append(subDirs, filepath.Join(dir, entry.Name()))
			continue
		}
		if isTerraplateFile(entry.Name()) {
			// Check that we haven't already detected a terrafile.
			// Multiple terrafiles are not allowed at this time.
			if terrafile != nil {
				return nil, fmt.Errorf("multiple terraplate files detected in folder %s", dir)
			}
			var (
				parseErr error
				path     = filepath.Join(dir, entry.Name())
			)
			terrafile, parseErr = ParseTerrafile(path)
			if parseErr != nil {
				return nil, fmt.Errorf("parsing terraplate file %s: %w", path, parseErr)
			}
			if ancestor != nil {
				ancestor.IsRoot = false
				terrafile.Ancestor = ancestor
			}

			terrafiles = append(terrafiles, terrafile)
		}
	}
	if terrafile == nil {
		terrafile = ancestor
	}
	for _, subDir := range subDirs {
		descFiles, err := walkDownDirectory(subDir, terrafile)
		if err != nil {
			return nil, err
		}
		terrafiles = append(terrafiles, descFiles...)
	}
	return terrafiles, nil
}

func walkUpDirectory(path string) (*Terrafile, error) {
	// Make sure we have an absolute path, as it's needed for filepath.Dir
	if !filepath.IsAbs(path) {
		var pathErr error
		path, pathErr = filepath.Abs(path)
		if pathErr != nil {
			return nil, fmt.Errorf("cannot get absolute path for %s: %w", path, pathErr)
		}
	}

	var (
		terrafile *Terrafile
		parentDir = filepath.Dir(path)
	)
	// Check if we cannot traverse any higher up. If so, return
	if path == parentDir {
		return terrafile, nil
	}

	entries, readErr := os.ReadDir(parentDir)
	if readErr != nil {
		return nil, fmt.Errorf("reading directory \"%s\": %w", parentDir, readErr)
	}
	for _, entry := range entries {
		if !entry.IsDir() && isTerraplateFile(entry.Name()) {
			// Check that we haven't already detected a terrafile.
			// Multiple terrafiles are not allowed at this time.
			if terrafile != nil {
				return nil, fmt.Errorf("multiple terraplate files detected in folder %s", parentDir)
			}
			var (
				parseErr error
				path     = filepath.Join(parentDir, entry.Name())
			)
			terrafile, parseErr = ParseTerrafile(path)
			if parseErr != nil {
				return nil, fmt.Errorf("parsing terraplate file %s: %w", path, parseErr)
			}
		}
	}

	ancestor, travErr := walkUpDirectory(parentDir)
	if travErr != nil {
		return nil, travErr
	}
	if ancestor != nil {
		// Ancestor is not a root module because it's in a parent directory
		ancestor.IsRoot = false
		// In case no terrafile was in this directory, terrafile is nil and should
		// be "replaced" with the ancestor
		if terrafile == nil {
			return ancestor, nil
		}
		// If terrafile is not nil, set the returned ancestor
		terrafile.Ancestor = ancestor
	}
	return terrafile, nil
}

func isTerraplateFile(name string) bool {
	return name == "terraplate.hcl" || strings.HasSuffix(name, ".tp.hcl")
}
