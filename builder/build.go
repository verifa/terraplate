package builder

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"text/template"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/verifa/terraplate/parser"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

// BuildData defines the data which is passed to the Go template engine
type BuildData struct {
	Locals    map[string]interface{}
	Variables map[string]interface{}
	Values    map[string]interface{}
	Terrafile *parser.Terrafile
	// RelativeDir is the relative directory from the root Terrafile to the
	// Terrafile being built
	RelativeDir string
	// RelativePath is the relative path from the root Terrafile to the Terrafile
	// being built
	RelativePath string
	// RelativeRootDir is the relative directory of the root Terrafile
	RelativeRootDir string
	// RootDir is the absolute directory of the root Terrafile
	RootDir string
}

// Build takes a TerraConfig as input and builds all the templates and terraform
// files
func Build(config *parser.TerraConfig) error {
	var buildErrs error
	for _, terrafile := range config.RootModules() {
		buildLocals, err := terrafile.LocalsAsGo()
		if err != nil {
			buildErrs = multierror.Append(buildErrs, fmt.Errorf("creating build locals: %w", err))
			continue
		}
		buildVars, err := terrafile.VariablesAsGo()
		if err != nil {
			buildErrs = multierror.Append(buildErrs, fmt.Errorf("creating build variables: %w", err))
			continue
		}
		buildValues, valuesErr := terrafile.ValuesAsGo()
		if valuesErr != nil {
			buildErrs = multierror.Append(buildErrs, fmt.Errorf("getting build values for terrafile \"%s\": %w", terrafile.Path, valuesErr))
			continue
		}
		buildData := BuildData{
			Locals:          buildLocals,
			Variables:       buildVars,
			Values:          buildValues,
			Terrafile:       terrafile,
			RelativePath:    terrafile.RelativePath(),
			RelativeDir:     terrafile.RelativeDir(),
			RelativeRootDir: terrafile.RelativeRootDir(),
			RootDir:         terrafile.RootDir(),
		}

		if err := buildTerraplate(terrafile, config, &buildData); err != nil {
			buildErrs = multierror.Append(buildErrs, fmt.Errorf("building Terraplate Terraform file: %w", err))
			continue
		}

		if err := buildTemplates(terrafile, buildData); err != nil {
			buildErrs = multierror.Append(buildErrs, fmt.Errorf("building templates: %w", err))
			continue
		}

		fmt.Printf("%s: Built %d templates\n", terrafile.RelativeDir(), len(terrafile.Templates))
	}

	if buildErrs != nil {
		return fmt.Errorf("building root modules: %s", buildErrs)
	}

	return nil
}

// buildTemplates builds the templates associated with the given terrafile
func buildTemplates(terrafile *parser.Terrafile, buildData BuildData) error {
	for _, tmpl := range terrafile.Templates {
		target := filepath.Join(terrafile.Dir, tmpl.Target)
		content := defaultTemplateHeader(terrafile, tmpl) + tmpl.Contents
		if err := templateWrite(buildData, tmpl.Name, content, target); err != nil {
			return fmt.Errorf("creating template %s in terrafile %s: %w", tmpl.Name, terrafile.RelativePath(), err)
		}
	}
	return nil
}

// buildTerraplate builds the terraplate terraform file which contains the
// variables (with defaults) and terraform block
func buildTerraplate(terrafile *parser.Terrafile, config *parser.TerraConfig, buildData *BuildData) error {
	path := filepath.Join(terrafile.Dir, "terraplate.tf")
	// Create the Terraform file
	tfFile := hclwrite.NewEmptyFile()

	// Write the terraform{} block
	tfBlock := hclwrite.NewBlock("terraform", []string{})
	// Set the required version if it was given
	if reqVer := terrafile.TerraformBlock.RequiredVersion; reqVer != "" {
		tfBlock.Body().SetAttributeValue("required_version", cty.StringVal(reqVer))
		tfBlock.Body().AppendNewline()
	}
	provBlock := hclwrite.NewBlock("required_providers", []string{})
	// We need to iterate over the required providers in order to avoid lots of
	// changes each time.
	// Iterate over the sorted keys and then extract the value for that key
	provMap := terrafile.TerraformBlock.RequiredProviders()
	for _, name := range sortedMapKeys(provMap) {
		value := provMap[name]
		ctyType, typeErr := gocty.ImpliedType(value)
		if typeErr != nil {
			return fmt.Errorf("implying required provider to cty type for provider %s: %w", name, typeErr)
		}
		ctyValue, ctyErr := gocty.ToCtyValue(value, ctyType)
		if ctyErr != nil {
			return fmt.Errorf("converting required provider to cty value for provider %s: %w", name, ctyErr)
		}
		provBlock.Body().SetAttributeValue(name, ctyValue)
	}
	tfBlock.Body().AppendBlock(provBlock)
	// If body is not empty, write the terraform block
	if isBodyEmpty(tfBlock.Body()) {
		tfFile.Body().AppendBlock(tfBlock)
		tfFile.Body().AppendNewline()
	}

	//
	// Write the locals {} block
	//
	localsMap := terrafile.Locals()
	localsBlock := hclwrite.NewBlock("locals", nil)
	for _, name := range sortedMapKeys(localsMap) {
		value := localsMap[name]
		localsBlock.Body().SetAttributeValue(name, value)
	}
	// If locals map is not empty, write the locals block to the terraplate file
	if len(localsMap) > 0 {
		tfFile.Body().AppendBlock(localsBlock)
		tfFile.Body().AppendNewline()
	}

	//
	// Write the variables {} block
	//
	// We need to iterate over the variables in order to avoid lots of
	// changes each time.
	// Iterate over the sorted keys and then extract the value for that key
	varMap := terrafile.Variables()
	for _, name := range sortedMapKeys(varMap) {
		value := varMap[name]
		varBlock := hclwrite.NewBlock("variable", []string{name})
		varBlock.Body().SetAttributeValue("default", value)
		tfFile.Body().AppendBlock(varBlock)
		tfFile.Body().AppendNewline()
	}

	// Create and write the file
	file, createErr := os.Create(path)
	if createErr != nil {
		return fmt.Errorf("creating file %s: %w", path, createErr)
	}

	contents := tfFile.Bytes()
	header := []byte(defaultTerraplateHeader(terrafile))
	contents = append(header, contents...)

	defer file.Close()
	if _, writeErr := file.Write(hclwrite.Format(hclwrite.Format(contents))); writeErr != nil {
		return fmt.Errorf("writing file %s: %w", path, writeErr)
	}
	return nil
}

func templateWrite(buildData BuildData, name string, text string, target string) error {
	tmpl, tmplErr := template.New(name).
		Option("missingkey=error").
		Funcs(builtinFuncs()).
		Parse(text)
	if tmplErr != nil {
		return tmplErr
	}
	// Add option to error on missing keys
	tmpl.Option("missingkey=error")

	var rawContents bytes.Buffer
	// Apply the template to the vars map and write the result to file.
	if execErr := tmpl.Execute(&rawContents, buildData); execErr != nil {
		return fmt.Errorf("executing template: %w", execErr)
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

func defaultTemplateHeader(tf *parser.Terrafile, tmpl *parser.TerraTemplate) string {
	return fmt.Sprintf(`#
	# NOTE: THIS FILE WAS AUTOMATICALLY GENERATED BY TERRAPLATE
	#
	# Terrafile: %s
	# Template: %s

	`, tf.RelativePath(), tmpl.Name)
}

func defaultTerraplateHeader(tf *parser.Terrafile) string {
	return fmt.Sprintf(`#
	# NOTE: THIS FILE WAS AUTOMATICALLY GENERATED BY TERRAPLATE
	#
	# Terrafile: %s

	`, tf.RelativePath())
}

// sortedMapKeys takes an input map and returns its keys sorted by alphabetical order
func sortedMapKeys(v interface{}) []string {
	rv := reflect.ValueOf(v)
	if rv.Type().Kind() != reflect.Map {
		panic(fmt.Sprintf("cannot sort map keys of non-map type %s", rv.Type().String()))
	}
	rvKeys := rv.MapKeys()
	keys := make([]string, 0, len(rvKeys))
	for _, key := range rvKeys {
		keys = append(keys, key.String())
	}
	sort.Strings(keys)
	return keys
}

func isBodyEmpty(body *hclwrite.Body) bool {
	return len(body.Attributes()) > 0 || len(body.Blocks()) > 0
}
