package builder

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/verifa/terraplate/parser"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

// Build takes a TerraConfig as input and builds all the templates and terraform
// files
func Build(config *parser.TerraConfig) error {
	var buildErrs error
	for _, terrafile := range config.RootModules() {

		if err := buildTerraplate(terrafile, config); err != nil {
			buildErrs = multierror.Append(buildErrs, fmt.Errorf("building Terraplate Terraform file: %w", err))
			continue
		}

		if err := buildTemplates(terrafile); err != nil {
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
func buildTemplates(tf *parser.Terrafile) error {
	for _, tmpl := range tf.Templates {
		data, dataErr := tf.BuildData()
		if dataErr != nil {
			return fmt.Errorf("getting build data for %s: %w", tf.Path, dataErr)
		}
		if tmpl.ConditionAttr != "" {
			condition, condErr := tmpl.Condition(data)
			if condErr != nil {
				return fmt.Errorf("evaluating condition for %s: %w", tf.Path, condErr)
			}
			if !condition {
				continue
			}
		}
		target := filepath.Join(tf.Dir, tmpl.Target)
		content := defaultTemplateHeader(tf, tmpl) + tmpl.Contents
		if err := parser.TemplateWrite(data, tmpl.Name, content, target); err != nil {
			return fmt.Errorf("creating template %s in terrafile %s: %w", tmpl.Name, tf.RelativePath(), err)
		}
	}
	return nil
}

// buildTerraplate builds the terraplate terraform file which contains the
// variables (with defaults) and terraform block
func buildTerraplate(terrafile *parser.Terrafile, config *parser.TerraConfig) error {
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
	if len(provMap) > 0 {
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
	}
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
