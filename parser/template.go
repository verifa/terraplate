package parser

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

// TerraTemplate defines the template{} block within a Terrafile
type TerraTemplate struct {
	Name     string `hcl:",label"`
	Contents string `hcl:"contents,attr"`
	// Target defines the target file to generate.
	// Defaults to the Name of the template with a ".tp.tf" extension
	Target string `hcl:"target,optional"`
	// ConditionAttr defines a string boolean that specifies whether this template
	// should be built or not.
	// The string can include Go templates, which means you can have dynamic
	// behaviour based on the Terrafile
	ConditionAttr string `hcl:"condition,optional"`
}

// Condition resolves the condition attribute to a boolean, or error.
// Errors can occur if either the templating errored or the conversion from string
// to bool is not possible.
func (t TerraTemplate) Condition(data *BuildData) (bool, error) {
	// If not set, the default is true (to build)
	if t.ConditionAttr == "" {
		return true, nil
	}
	// First tempalte it
	contents, execErr := ExecTemplate(data, "condition", t.ConditionAttr)
	if execErr != nil {
		return false, fmt.Errorf("templating condition for template %s: %w", t.Name, execErr)
	}
	condition, parseErr := strconv.ParseBool(contents.String())
	if parseErr != nil {
		return false, fmt.Errorf("converting condition string to bool for template %s: %w", t.Name, parseErr)
	}
	return condition, nil
}

func TemplateWrite(buildData *BuildData, name string, text string, target string) error {
	rawContents, execErr := ExecTemplate(buildData, name, text)
	if execErr != nil {
		return execErr
	}
	// Format the contents to make it nice HCL
	formattedContents := hclwrite.Format(rawContents.Bytes())

	file, createErr := os.Create(target)
	if createErr != nil {
		return fmt.Errorf("creating file %s: %w", target, createErr)
	}
	defer file.Close()
	if _, writeErr := file.Write(formattedContents); writeErr != nil {
		return fmt.Errorf("writing file %s: %w", target, writeErr)
	}
	return nil
}

func ExecTemplate(buildData *BuildData, name string, text string) (*bytes.Buffer, error) {
	tmpl, tmplErr := commonTemplate(name).Parse(text)
	if tmplErr != nil {
		return nil, tmplErr
	}

	var contents bytes.Buffer
	if execErr := tmpl.Execute(&contents, buildData); execErr != nil {
		return nil, fmt.Errorf("executing template: %w", execErr)
	}
	return &contents, nil
}

func commonTemplate(name string) *template.Template {
	return template.New(name).
		Option("missingkey=error").
		Funcs(sprig.TxtFuncMap())
}
