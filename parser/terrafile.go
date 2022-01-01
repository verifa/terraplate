package parser

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2/hclsimple"
	"github.com/imdario/mergo"
	"github.com/zclconf/go-cty/cty"
	ctyjson "github.com/zclconf/go-cty/cty/json"
)

type Terrafile struct {
	Path string `hcl:"-"`
	Dir  string `hcl:"-"`
	// IsRoot tells whether this terrafile is for a root module
	IsRoot bool `hcl:"-"`
	// Templates defines the list of templates that this Terrafile defines
	Templates []*TerraTemplate `hcl:"template,block"`
	Variables *TerraVariables  `hcl:"variables,block"`
	Values    *TerraValues     `hcl:"values,block"`

	RequiredVersion   string                  `hcl:"required_version,optional"`
	RequiredProviders *TerraRequiredProviders `hcl:"required_providers,block"`

	// Variables map[string]cty.Value `hcl:",remain"`
	// Ancestor defines any parent/ancestor Terrafiles that this Terrafile
	// should inherit from
	Ancestor *Terrafile `hcl:"-"`
}

type TerraValues struct {
	Values map[string]cty.Value `hcl:",remain"`
}

type TerraVariables struct {
	Variables map[string]cty.Value `hcl:",remain"`
}

type TerraRequiredProviders struct {
	RequiredProviders map[string]RequiredProvider `hcl:",remain"`
}

type RequiredProvider struct {
	Source  string `hcl:"source,attr" cty:"source"`
	Version string `hcl:"version,attr" cty:"version"`
}

type TerraTemplate struct {
	Name string `hcl:",label"`
	// Source defines the source template filename
	Source string `hcl:"source,optional"`
	// Target defines the target file to generate.
	// Defaults to Source and removing any .tmpl extensions
	Target string `hcl:"target,optional"`
	// Build specifies whether this should be built or not (defaults to true)
	Build *bool `hcl:"build,optional"`

	// TemplateDir is the directory containing the templates.
	// Joining the paths TemplateDir and Source should point to a valid template
	TemplateDir string
	// Ancestors defines any parent/ancestor templates with the same name
	// that should be merged/overwritten
	Ancestors []*TerraTemplate
}

func ParseTerrafile(file string) (*Terrafile, error) {

	var terrafile Terrafile
	if err := hclsimple.DecodeFile(file, nil, &terrafile); err != nil {
		return nil, fmt.Errorf("decoding terraplate file %s: %w", file, err)
	}
	terrafile.Path = file
	terrafile.Dir = filepath.Dir(file)
	// Set the default to be a root module. If an ancestor is added it is set to false
	terrafile.IsRoot = true

	// Check if there's a template directory
	tmplDir := filepath.Join(filepath.Dir(file), "templates")

	if tmplErr := parseTemplates(&terrafile, tmplDir); tmplErr != nil {
		return nil, fmt.Errorf("parsing templates: %w", tmplErr)
	}

	return &terrafile, nil
}

func parseTemplates(terrafile *Terrafile, dir string) error {

	entries, readErr := os.ReadDir(dir)
	if readErr != nil {
		if os.IsNotExist(readErr) {
			// This is fine, and there's nothing to do
			return nil
		}
		return fmt.Errorf("reading templates directory %s: %w", dir, readErr)
	}
	var tmplSourceMap = make(map[string]*TerraTemplate)
	var tmplNameMap = make(map[string]*TerraTemplate)
	var defaultBuild = true
	for _, tmpl := range terrafile.Templates {
		// Set the defaults for defined templates
		if tmpl.Build == nil {
			tmpl.Build = &defaultBuild
		}
		if tmpl.Source != "" {
			tmplSourceMap[tmpl.Source] = tmpl
		}
		tmplNameMap[tmpl.Name] = tmpl
		tmpl.TemplateDir = dir
	}

	for _, entry := range entries {
		if entry.IsDir() {
			// Skip directories
			continue
		}
		if strings.HasSuffix(entry.Name(), ".tf") || strings.HasSuffix(entry.Name(), ".tf.tmpl") {
			// Trim the tmpl suffix to get the end terraform filename
			tfFile := strings.TrimSuffix(entry.Name(), ".tmpl")
			defaultName := strings.TrimSuffix(tfFile, ".tf")
			// Check if the template was defined in the terraplate configuration.
			// This can be done either by source file matching or the default
			// name matching
			if tmpl, ok := tmplNameMap[defaultName]; ok {
				if tmpl.Source == entry.Name() {
					if tmpl.Target == "" {
						tmpl.Target = tfFile
					}
					continue
				}
				// If no source, set it
				if tmpl.Source == "" {
					tmpl.Source = entry.Name()
					if tmpl.Target == "" {
						tmpl.Target = tfFile
					}
					continue
				}
				// Otherwise the names match but the source is different.
				// Probably an error, inform the user
				return fmt.Errorf("template with same name as file detected but the sources do not match for template %s", tmpl.Name)
			}
			if _, ok := tmplSourceMap[entry.Name()]; ok {
				// If the default name was not the same but the sources match, then reuse the template
				continue
			}
			// If the detected terraform file was not declared in the terraplate configuration,
			// create the template automatically
			var tmpl TerraTemplate
			tmpl.Name = defaultName
			tmpl.Source = entry.Name()
			tmpl.Target = tfFile
			tmpl.TemplateDir = dir
			tmpl.Build = &defaultBuild

			terrafile.Templates = append(terrafile.Templates, &tmpl)
		}
	}
	return nil
}

func (t *Terrafile) TraverseAncestors(visit func(ancestor *Terrafile) error) error {
	// Recurse through all ancestors and add the templates to the template map
	var ancestor = t.Ancestor
	for ancestor != nil {
		if err := visit(ancestor); err != nil {
			return err
		}
		// Set next ancestor
		ancestor = ancestor.Ancestor
	}
	return nil
}

// HasSource returns true if the template has a source (or an ancestor has one)
// and false if not. In which case, there's likely something wrong
func (t *TerraTemplate) HasSource() bool {
	if t.Source != "" {
		return true
	}
	for _, an := range t.Ancestors {
		if an.Source != "" {
			return true
		}
	}
	return false
}

// SourceFiles returns the source files for a template and it's ancestors, starting
// with the highest ancestor, and ending with the template itself (if it has a source)
func (t *TerraTemplate) SourceFiles() []string {
	var sourceFiles []string
	// Iterate in reverse order
	for i := len(t.Ancestors) - 1; i >= 0; i-- {
		source := t.Ancestors[i].Source
		if source != "" {
			sourceFiles = append(sourceFiles, filepath.Join(t.Ancestors[i].TemplateDir, source))
		}
	}
	if t.Source != "" {
		sourceFiles = append(sourceFiles, filepath.Join(t.TemplateDir, t.Source))
	}
	return sourceFiles
}
func (t *TerraTemplate) BuildTarget() string {
	if t.Target != "" {
		return t.Target
	}
	for _, an := range t.Ancestors {
		if an.Target != "" {
			return an.Target
		}
	}
	return ""
}

func (t *Terrafile) BuildVariables() map[string]cty.Value {
	var buildVars map[string]cty.Value
	if t.Variables != nil {
		buildVars = t.Variables.Variables
	} else {
		buildVars = make(map[string]cty.Value)
	}
	if t.Ancestor == nil {
		return buildVars
	}
	for name, value := range t.Ancestor.BuildVariables() {
		if _, ok := buildVars[name]; !ok {
			buildVars[name] = value
		}
	}
	return buildVars
}

func (t *Terrafile) BuildValues() (map[string]interface{}, error) {
	var buildValues = make(map[string]interface{})
	if t.Values != nil {
		for name, value := range t.Values.Values {
			// This is a bit HACKY, but there is not a straight forward way to
			// get cty values into interface{} values, but we can JSON serialize
			// and then back to interface{}
			simple := ctyjson.SimpleJSONValue{
				Value: value,
			}
			b, err := simple.MarshalJSON()
			if err != nil {
				return nil, fmt.Errorf("json marshalling cty value %s: %w", name, err)
			}

			var val interface{}
			if err := json.Unmarshal(b, &val); err != nil {
				return nil, fmt.Errorf("json unmarshalling cty value %s: %w", name, err)
			}

			buildValues[name] = val
		}
	}
	if t.Ancestor == nil {
		return buildValues, nil
	}
	anBuildValues, err := t.Ancestor.BuildValues()
	if err != nil {
		return nil, err
	}
	for name, anVal := range anBuildValues {
		if curVal, ok := buildValues[name]; ok {
			// If curVal is a map, then we should merge it, if the ancestor value
			// is also a map
			if dstMap, ok := curVal.(map[string]interface{}); ok {
				if srcMap, ok := anVal.(map[string]interface{}); ok {
					if mergeErr := mergo.Merge(&dstMap, srcMap); mergeErr != nil {
						return nil, fmt.Errorf("merging values for terrafile %s: %w", t.Path, mergeErr)
					}
					buildValues[name] = dstMap
					continue
				}
			}
		} else {
			// If it did not merge, then simply overwrite
			buildValues[name] = anVal
		}
	}
	return buildValues, nil
}

func (t *Terrafile) BuildRequiredProviders() map[string]RequiredProvider {
	var reqProv map[string]RequiredProvider
	if t.RequiredProviders != nil {
		reqProv = t.RequiredProviders.RequiredProviders
	} else {
		reqProv = make(map[string]RequiredProvider)
	}
	if t.Ancestor == nil {
		return reqProv
	}
	for name, value := range t.Ancestor.BuildRequiredProviders() {
		if _, ok := reqProv[name]; !ok {
			reqProv[name] = value
		}
	}
	return reqProv
}

func (t *Terrafile) BuildRequiredVersion() string {
	var requiredVersion = t.RequiredVersion
	t.TraverseAncestors(func(ancestor *Terrafile) error {
		if requiredVersion == "" {
			if ancestor.RequiredVersion != "" {
				requiredVersion = ancestor.RequiredVersion
			}
		}
		return nil
	})
	return requiredVersion
}

func (t *Terrafile) BuildTemplates() []*TerraTemplate {
	var tmplMap = make(map[string]*TerraTemplate)

	for _, tmpl := range t.Templates {
		tmplMap[tmpl.Name] = tmpl
	}

	// var ancestor = t.Ancestor
	// for ancestor != nil {
	// 	for _, tmpl := range ancestor.Templates {
	// 		t, ok := tmplMap[tmpl.Name]
	// 		if ok {
	// 			t.Ancestors = append(t.Ancestors, tmpl)
	// 			continue
	// 		}

	// 		// If it did not exist, then set it
	// 		tmplMap[tmpl.Name] = tmpl
	// 	}
	// 	// Set next ancestor
	// 	ancestor = ancestor.Ancestor
	// }
	// Recurse through all ancestors and add the templates to the template map
	t.TraverseAncestors(func(ancestor *Terrafile) error {
		for _, tmpl := range ancestor.Templates {
			t, ok := tmplMap[tmpl.Name]
			if ok {
				t.Ancestors = append(t.Ancestors, tmpl)
				continue
			}
			// If it did not exist, then set it
			tmplMap[tmpl.Name] = tmpl
		}
		return nil
	})

	var templates = make([]*TerraTemplate, 0, len(tmplMap))
	for _, tmpl := range tmplMap {
		// Check whether this template should be built, if not part
		if !*tmpl.Build {
			continue
		}
		templates = append(templates, tmpl)
	}
	return templates
}
