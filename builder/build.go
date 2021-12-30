package builder

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"text/template"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/verifa/terraplate/parser"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

// BuildData defines the data which is passed to the Go template engine
type BuildData struct {
	Variables map[string]cty.Value
	Values    map[string]interface{}
	Terrafile *parser.Terrafile
}

func Build(config *parser.TerraConfig) error {
	for _, terrafile := range config.BuildFiles() {
		buildValues, valuesErr := terrafile.BuildValues()
		if valuesErr != nil {
			return fmt.Errorf("getting build values for terrafile \"%s\": %w", terrafile.Path, valuesErr)
		}
		buildData := BuildData{
			Variables: terrafile.BuildVariables(),
			Values:    buildValues,
			Terrafile: terrafile,
		}

		if err := buildTerraplate(terrafile, config, &buildData); err != nil {
			return fmt.Errorf("building Terraplate Terraform file: %w", err)
		}

		baseDir := filepath.Dir(terrafile.Path)
		for _, terraTmpl := range terrafile.BuildTemplates() {
			// Make sure the template has at least one source
			if !terraTmpl.HasSource() {
				return fmt.Errorf("no source found in template or ancestors for \"%s\" in terraplate file \"%s\"", terraTmpl.Name, terrafile.Path)
			}

			var tmplBytes []byte
			// Iterate over source files and combine into bytes
			for _, src := range terraTmpl.SourceFiles() {
				file, openErr := os.Open(src)
				if openErr != nil {
					return fmt.Errorf("opening template file %s: %w", src, openErr)
				}
				defer file.Close()
				contents, readErr := io.ReadAll(file)
				if readErr != nil {
					return fmt.Errorf("reading template file %s: %w", src, readErr)
				}
				tmplBytes = append(tmplBytes, contents...)
			}

			tmpl, tmplErr := template.New(terraTmpl.Target).
				Option("missingkey=error").
				Funcs(builtinFuncs()).
				Parse(string(tmplBytes))
			if tmplErr != nil {
				return fmt.Errorf("parsing templates for terraplate file %s: %w", terrafile.Path, tmplErr)
			}
			// Add option to error on missing keys
			tmpl.Option("missingkey=error")

			target := filepath.Join(baseDir, terraTmpl.BuildTarget())
			var contents bytes.Buffer
			// Apply the template to the vars map and write the result to file.
			if execErr := tmpl.Execute(&contents, buildData); execErr != nil {
				return fmt.Errorf("executing template %s: %w", terrafile.Path, execErr)
			}
			file, createErr := os.Create(target)
			if createErr != nil {
				return fmt.Errorf("creating file %s: %w", target, createErr)
			}
			defer file.Close()
			if _, writeErr := file.Write(contents.Bytes()); writeErr != nil {
				return fmt.Errorf("writing file %s: %w", target, writeErr)
			}

		}

		// Generate the tfvars file
		tfvars := filepath.Join(baseDir, "terraform.tfvars")
		file, createErr := os.Create(tfvars)
		if createErr != nil {
			return fmt.Errorf("creating file %s: %w", tfvars, createErr)
		}
		defer file.Close()

		tmpl, tmplErr := template.New("terraform.tfvars").Funcs(builtinFuncs()).Option("missingkey=error").Parse(tfvarsTemplate)
		if tmplErr != nil {
			return fmt.Errorf("parsing tfvars for terraplate file %s: %w", terrafile.Path, tmplErr)
		}
		var tfvarsBuffer bytes.Buffer
		// Apply the template to the vars map and write the result to file.
		if execErr := tmpl.Execute(&tfvarsBuffer, buildData); execErr != nil {
			return fmt.Errorf("executing template %s: %w", terrafile.Path, execErr)
		}
		if _, writeErr := file.Write(hclwrite.Format(tfvarsBuffer.Bytes())); writeErr != nil {
			return fmt.Errorf("writing tfvars file %s: %w", tfvars, writeErr)
		}

		fmt.Println("Successfully built Terraplate in", terrafile.Dir)
	}

	return nil
}

// buildTerraplate builds the terraplate terraform file which contains the
// variables (with defaults) and terraform block
func buildTerraplate(terrafile *parser.Terrafile, config *parser.TerraConfig, buildData *BuildData) error {
	baseDir := filepath.Dir(terrafile.Path)
	path := filepath.Join(baseDir, "terraplate.tf")
	// Create the Terraform file
	tfFile := hclwrite.NewEmptyFile()

	// Write the terraform{} block
	tfBlock := hclwrite.NewBlock("terraform", []string{})
	// Set the required version if it was given
	if reqVer := terrafile.BuildRequiredVersion(); reqVer != "" {
		tfBlock.Body().SetAttributeValue("required_version", cty.StringVal(reqVer))
		tfBlock.Body().AppendNewline()
	}
	provBlock := hclwrite.NewBlock("required_providers", []string{})
	for name, value := range terrafile.BuildRequiredProviders() {
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
	tfFile.Body().AppendBlock(tfBlock)
	tfFile.Body().AppendNewline()

	// Write variables
	for name, value := range buildData.Variables {
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
	defer file.Close()
	if _, writeErr := file.Write(hclwrite.Format(hclwrite.Format(tfFile.Bytes()))); writeErr != nil {
		return fmt.Errorf("writing file %s: %w", path, writeErr)
	}
	return nil
}