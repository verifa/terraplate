package parser

import (
	"bytes"
	"fmt"
	"os"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

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
