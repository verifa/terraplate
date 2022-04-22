package parser

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func DefaultConfig() *Config {
	return &Config{
		Chdir: ".",
	}
}

type Config struct {
	Chdir string
}

func Parse(config *Config) (*TerraConfig, error) {
	// Check that the directory exists and is a directory
	dirStat, statErr := os.Stat(config.Chdir)
	if statErr != nil {
		return nil, statErr
	}
	if !dirStat.IsDir() {
		return nil, fmt.Errorf("given directory is not a directory: %s", config.Chdir)
	}

	ancestor, travErr := walkUpDirectory(config.Chdir)
	if travErr != nil {
		return nil, fmt.Errorf("looking for parent terraplate.hcl files: %w", travErr)
	}

	terrafiles, walkErr := walkDownDirectory(config.Chdir, ancestor)
	if walkErr != nil {
		return nil, fmt.Errorf("looking for terraplate.hcl files: %w", walkErr)
	}

	// Check if any terrafiles were found. If not, return an error
	if len(terrafiles) == 0 {
		return nil, errors.New("no terraplate files found")
	}

	tfc := TerraConfig{
		Terrafiles: terrafiles,
	}

	// Terrafiles inherit values from ancestors. Let's resolve the root modules
	// so that they are ready for building/executing
	if err := tfc.MergeTerrafiles(); err != nil {
		return nil, fmt.Errorf("resolving inheritance: %w", err)
	}

	return &tfc, nil
}

func traverseUpDirectory(path string, visit func(dir string) (bool, error)) error {
	if !filepath.IsAbs(path) {
		var pathErr error
		path, pathErr = filepath.Abs(path)
		if pathErr != nil {
			return fmt.Errorf("cannot get absolute path for %s: %w", path, pathErr)
		}
	}

	fileInfo, statErr := os.Stat(path)
	if statErr != nil {
		return statErr
	}
	if !fileInfo.IsDir() {
		path = filepath.Dir(path)
	}

	// Check if we cannot traverse any higher up. If so, return
	for path != filepath.Dir(path) {
		proceed, err := visit(path)
		if err != nil {
			return err
		}
		// If false was returned, indicating not to proceed, then return
		if !proceed {
			return nil
		}
		path = filepath.Dir(path)
	}

	return nil
}

func walkUpDirectory(path string) (*Terrafile, error) {
	var (
		skipFirst      = false
		childTerrafile *Terrafile
	)
	travErr := traverseUpDirectory(path, func(dir string) (bool, error) {
		// Skip the first directory as it will get processed when we walk down
		// the directory structure
		if !skipFirst {
			skipFirst = true
			return true, nil
		}

		entries, readErr := os.ReadDir(dir)
		if readErr != nil {
			return false, fmt.Errorf("reading directory \"%s\": %w", dir, readErr)
		}

		var terrafile *Terrafile
		// Iterate over files and check if any of them are Terrafiles.
		// If they are, parse them. If there's multiple, it's an error (for now).
		for _, entry := range entries {
			if !entry.IsDir() && isTerraplateFile(entry.Name()) {
				// Check that we haven't already detected a terrafile.
				// Multiple terrafiles are not allowed at this time.
				if terrafile != nil {
					return false, fmt.Errorf("multiple terraplate files detected in folder %s", dir)
				}
				var (
					parseErr error
					path     = filepath.Join(dir, entry.Name())
				)
				terrafile, parseErr = parseTerrafile(path)
				if parseErr != nil {
					return false, fmt.Errorf("parsing terraplate file %s: %w", path, parseErr)
				}
			}
		}

		// If no terrafile was found, just continue traversing up
		if terrafile == nil {
			return true, nil
		}
		if childTerrafile != nil {
			// Terrafile is not a root module because it has a child
			terrafile.IsRoot = false
			// Create parent/child relationship
			terrafile.Children = append(terrafile.Children, childTerrafile)
			childTerrafile.Ancestor = terrafile
		}
		// Set current terrafile as the new child terrafile for the next traversal
		childTerrafile = terrafile

		return true, nil
	})
	if travErr != nil {
		return nil, fmt.Errorf("walking up directories: %w", travErr)
	}
	if childTerrafile != nil {
		rootMods := childTerrafile.rootModules()
		if len(rootMods) == 1 {
			return rootMods[0], nil
		}

		return nil, fmt.Errorf("unexpected: terrafile has more than one root module after walking up directories, total %d", len(rootMods))
	}
	// Returning nil is ok. It means we did not find any terrafiles but also did
	// not encounter an error
	return nil, nil
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
			terrafile, parseErr = parseTerrafile(path)
			if parseErr != nil {
				return nil, fmt.Errorf("parsing terraplate file %s: %w", path, parseErr)
			}
			if ancestor != nil {
				ancestor.IsRoot = false
				terrafile.Ancestor = ancestor
				ancestor.Children = append(ancestor.Children, terrafile)
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

func isTerraplateFile(name string) bool {
	return name == "terraplate.hcl" || strings.HasSuffix(name, ".tp.hcl")
}
