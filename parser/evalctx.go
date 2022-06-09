package parser

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
)

func evalCtx(dir string) *hcl.EvalContext {
	return &hcl.EvalContext{
		Functions: map[string]function.Function{
			"read_template": readTemplateFunc(dir),
		},
	}
}

// readTemplateFunc creates an HCL function that will read the contents of a
// template file by the given name, starting at the directory provided.
// It will first check for the template file within a "templates" directory
// (if it exists), and then in the root of the given directory.
// It will traverse up directories until it finds a template with that name
// and return the contents of the first match that it finds.
// If no template is found, it returns an error
func readTemplateFunc(dir string) function.Function {
	return function.New(&function.Spec{
		Params: []function.Parameter{
			{
				Name: "file",
				Type: cty.String,
			},
		},
		Type: function.StaticReturnType(cty.String),
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			file := args[0].AsString()

			var (
				contents string
				found    bool
				readErr  error
			)
			travErr := TraverseUpDirectory(dir, func(travDir string) (bool, error) {
				var path string
				// Check if there is a template file first in the templates
				// directory, and then in the directory we are traversing
				path = filepath.Join(travDir, "templates", file)
				contents, found, readErr = readTemplate(path)
				if readErr != nil {
					return false, readErr
				}
				if found {
					// Indicate not to continue traversing
					return false, nil
				}

				// Try again without the templates directory
				path = filepath.Join(travDir, file)
				contents, found, readErr = readTemplate(path)
				if readErr != nil {
					return false, readErr
				}
				if found {
					// Indicate not to continue traversing
					return false, nil
				}
				// If not found, and no errors, continue traversing
				return true, nil
			})
			if travErr != nil {
				return cty.NilVal, fmt.Errorf("")
			}

			if !found {
				return cty.NilVal, fmt.Errorf("could not find template %s", file)
			}

			return cty.StringVal(contents), nil
		},
	})
}

func readTemplate(path string) (string, bool, error) {
	bytes, readErr := os.ReadFile(path)
	if readErr != nil {
		if !os.IsNotExist(readErr) {
			return "", false, fmt.Errorf("reading file %s: %w", path, readErr)
		}
		return "", false, nil
	}
	return string(bytes), true, nil
}
