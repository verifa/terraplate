package parser

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/hashicorp/hcl/v2/hclsimple"
	"github.com/imdario/mergo"
	"github.com/zclconf/go-cty/cty"
	ctyjson "github.com/zclconf/go-cty/cty/json"
)

// defaultTerrafile sets default values for a Terrafile that are used when
// parsing a new Terrafile
var defaultTerrafile = Terrafile{
	// BuildBlock: &BuildBlock{},
	ExecBlock: &ExecBlock{
		PlanBlock: &ExecPlanBlock{
			Input:   false,
			Lock:    true,
			Out:     "tfplan",
			SkipOut: false,
		},
	},
	TerraformBlock: &TerraformBlock{
		RequiredProvidersBlock: &TerraRequiredProviders{
			RequiredProviders: make(map[string]RequiredProvider),
		},
	},
}

type Terrafile struct {
	// Path
	Path string
	Dir  string
	// IsRoot tells whether this terrafile is for a root module
	IsRoot bool
	// Templates defines the list of templates that this Terrafile defines
	Templates []*TerraTemplate `hcl:"template,block"`

	LocalsBlock    *TerraLocals    `hcl:"locals,block"`
	VariablesBlock *TerraVariables `hcl:"variables,block"`
	ValuesBlock    *TerraValues    `hcl:"values,block"`

	TerraformBlock *TerraformBlock `hcl:"terraform,block"`

	// BuildBlock *BuildBlock `hcl:"build,block"`
	ExecBlock *ExecBlock `hcl:"exec,block"`

	// Ancestor defines any parent/ancestor Terrafiles that this Terrafile
	// should inherit from
	Ancestor *Terrafile

	// Children contains any child Terrafiles to this Terrafile
	Children []*Terrafile
}

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

// TerraformBlock defines the terraform{} block within a Terrafile
type TerraformBlock struct {
	RequiredVersion        string                  `hcl:"required_version,optional"`
	RequiredProvidersBlock *TerraRequiredProviders `hcl:"required_providers,block"`
}

// RequiredProviders returns the map of terraform required_providers, or nil
func (tb *TerraformBlock) RequiredProviders() map[string]RequiredProvider {
	if tb.RequiredProvidersBlock == nil {
		return nil
	}
	return tb.RequiredProvidersBlock.RequiredProviders
}

// TerraRequiredProviders defines the map of required_providers
type TerraRequiredProviders struct {
	RequiredProviders map[string]RequiredProvider `hcl:",remain"`
}

// RequiredProvider defines the body of required_providers
type RequiredProvider struct {
	Source  string `hcl:"source,attr" cty:"source"`
	Version string `hcl:"version,attr" cty:"version"`
}

type TerraLocals struct {
	Locals map[string]cty.Value `hcl:",remain"`
}

type TerraValues struct {
	Values map[string]cty.Value `hcl:",remain"`
}

type TerraVariables struct {
	Variables map[string]cty.Value `hcl:",remain"`
}

// parseTerrafile parses the terrafile given by the input string and returns
// the Terrafile or an error if something went wrong
func parseTerrafile(file string) (*Terrafile, error) {

	terrafileDir := filepath.Dir(file)

	var terrafile Terrafile
	if err := hclsimple.DecodeFile(file, evalCtx(terrafileDir), &terrafile); err != nil {
		return nil, fmt.Errorf("decoding terraplate file %s: %w", file, err)
	}
	terrafile.Path = file
	terrafile.Dir = terrafileDir
	// Set the default to be a root module. If an ancestor is added it is set to false
	terrafile.IsRoot = true

	for _, tmpl := range terrafile.Templates {
		// Set the defaults for defined templates
		if tmpl.Target == "" {
			tmpl.Target = tmpl.Name + ".tp.tf"
		}
	}

	return &terrafile, nil
}

// traverseChildren goes down the tree of terrafiles calling the visit function
// on each terrafile in the path
func (t *Terrafile) traverseChildren(visit func(parent *Terrafile, tf *Terrafile) error) error {
	for _, child := range t.Children {
		if err := visit(t, child); err != nil {
			return err
		}
		if err := child.traverseChildren(visit); err != nil {
			return err
		}
	}
	return nil
}

// traverseAncestors goes up the tree of ancestors calling the visit function
// on each terrafile in the path
func (t *Terrafile) traverseAncestors(visit func(ancestor *Terrafile) error) error {
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

// traverseAncestorsReverse goes down the tree of ancestors calling the visit function
// on each terrafile in the path
func (t *Terrafile) traverseAncestorsReverse(visit func(ancestor *Terrafile) error) error {
	// Traverse up the tree first and make a reversed list of the terrafiles in
	// the path, so that we can iterate over the terrafiles in reverse order
	var path = make([]*Terrafile, 0)
	t.traverseAncestors(func(ancestor *Terrafile) error {
		path = append([]*Terrafile{ancestor}, path...)
		return nil
	})
	for _, tf := range path {
		if err := visit(tf); err != nil {
			return err
		}
	}
	return nil
}

// rootAncestor returns the upmost (root) ancestor in the hierarchy
func (t *Terrafile) rootAncestor() *Terrafile {
	tf := t
	t.traverseAncestors(func(ancestor *Terrafile) error {
		tf = ancestor
		return nil
	})
	return tf
}

// rootModules returns the root modules that are children of this terrafile
func (t *Terrafile) rootModules() []*Terrafile {
	var terrafiles = make([]*Terrafile, 0)
	if t.IsRoot {
		terrafiles = append(terrafiles, t)
	}
	t.traverseChildren(func(parent *Terrafile, child *Terrafile) error {
		if child.IsRoot {
			terrafiles = append(terrafiles, child)
		}
		return nil
	})
	return terrafiles
}

func (t *Terrafile) BuildData() (*BuildData, error) {
	buildLocals, err := t.LocalsAsGo()
	if err != nil {
		return nil, fmt.Errorf("creating build locals: %w", err)
	}
	buildVars, err := t.VariablesAsGo()
	if err != nil {
		return nil, fmt.Errorf("creating build variables: %w", err)
	}
	buildValues, valuesErr := t.ValuesAsGo()
	if valuesErr != nil {
		return nil, fmt.Errorf("getting build values for terrafile \"%s\": %w", t.Path, valuesErr)
	}
	return &BuildData{
		Locals:          buildLocals,
		Variables:       buildVars,
		Values:          buildValues,
		Terrafile:       t,
		RelativePath:    t.RelativePath(),
		RelativeDir:     t.RelativeDir(),
		RelativeRootDir: t.RelativeRootDir(),
		RootDir:         t.RootDir(),
	}, nil
}

// RelativeRootDir returns the relative directory of the root Terrafile
func (t *Terrafile) RelativeRootDir() string {
	root := t.rootAncestor()
	rootAbsDir, absErr := filepath.Abs(root.Dir)
	if absErr != nil {
		panic(fmt.Sprintf("cannot get absolute path to Terrafile %s: %s", root.Path, absErr.Error()))
	}
	tfAbsDir, absErr := filepath.Abs(t.Dir)
	if absErr != nil {
		panic(fmt.Sprintf("cannot get absolute path to Terrafile %s: %s", t.Path, absErr.Error()))
	}
	relPath, relErr := filepath.Rel(tfAbsDir, rootAbsDir)
	if relErr != nil {
		panic(fmt.Sprintf("cannot get relative path from Terrafile %s to %s: %s", t.Path, root.Path, relErr.Error()))
	}
	return relPath
}

// RootDir returns the absolute directory of the root Terrafile
func (t *Terrafile) RootDir() string {
	root := t.rootAncestor()
	rootAbsDir, absErr := filepath.Abs(root.Dir)
	if absErr != nil {
		panic(fmt.Sprintf("cannot get absolute path to Terrafile %s: %s", root.Path, absErr.Error()))
	}
	return rootAbsDir
}

// RelativeDir gets the relative directory from the Root Terrafile to this Terrafile
func (t *Terrafile) RelativeDir() string {
	root := t.rootAncestor()
	rootAbsDir, absErr := filepath.Abs(root.Dir)
	if absErr != nil {
		panic(fmt.Sprintf("cannot get absolute path to Terrafile %s: %s", root.Path, absErr.Error()))
	}
	tfAbsDir, absErr := filepath.Abs(t.Dir)
	if absErr != nil {
		panic(fmt.Sprintf("cannot get absolute path to Terrafile %s: %s", t.Path, absErr.Error()))
	}
	relPath, relErr := filepath.Rel(rootAbsDir, tfAbsDir)
	if relErr != nil {
		panic(fmt.Sprintf("cannot get relative path from ancestor Terrafile %s to %s: %s", root.Path, t.Path, relErr.Error()))
	}
	return relPath
}

// RelativePath gets the relative directory from the Root Terrafile to this Terrafile
func (t *Terrafile) RelativePath() string {
	root := t.rootAncestor()
	rootAbsDir, absErr := filepath.Abs(root.Dir)
	if absErr != nil {
		panic(fmt.Sprintf("cannot get absolute path to Terrafile %s: %s", root.Path, absErr.Error()))
	}
	tfAbsPath, absErr := filepath.Abs(t.Path)
	if absErr != nil {
		panic(fmt.Sprintf("cannot get absolute path to Terrafile %s: %s", t.Path, absErr.Error()))
	}
	relPath, relErr := filepath.Rel(rootAbsDir, tfAbsPath)
	if relErr != nil {
		panic(fmt.Sprintf("cannot get relative path from ancestor Terrafile %s to %s: %s", root.Path, t.Path, relErr.Error()))
	}
	return relPath
}

// mergeTerrafiles will merge a terrafile and it's direct ancestor.
// We cannot simply merge all fields as templates (for example) are treated a bit
// special
func (t *Terrafile) mergeTerrafile(parent *Terrafile) error {

	t.mergeLocals(parent)
	t.mergeVariables(parent)
	t.mergeValues(parent)
	t.mergeTemplates(parent)
	t.mergeTerraformBlock(parent)

	if mergeErr := t.mergeExecBlock(parent); mergeErr != nil {
		return mergeErr
	}

	return nil
}

func (t *Terrafile) mergeExecBlock(parent *Terrafile) error {
	// If terrafile's exec block is nil, we can simply inherit the ancestor's one
	if t.ExecBlock == nil {
		t.ExecBlock = &ExecBlock{}
	}
	// Merge ExecBlock
	if err := mergo.Merge(t.ExecBlock, parent.ExecBlock); err != nil {
		return fmt.Errorf("merging exec{} block: %w", err)
	}
	return nil
}

func (t *Terrafile) mergeLocals(parent *Terrafile) {
	tfLocals := t.Locals()
	if t.Locals() == nil {
		tfLocals = make(map[string]cty.Value)
	}
	for name, value := range parent.Locals() {
		if _, ok := tfLocals[name]; !ok {
			tfLocals[name] = value
		}
	}
	t.LocalsBlock = &TerraLocals{
		Locals: tfLocals,
	}

}

func (t *Terrafile) mergeVariables(parent *Terrafile) {
	tfVars := t.Variables()
	if t.Variables() == nil {
		tfVars = make(map[string]cty.Value)
	}
	for name, value := range parent.Variables() {
		if _, ok := tfVars[name]; !ok {
			tfVars[name] = value
		}
	}
	t.VariablesBlock = &TerraVariables{
		Variables: tfVars,
	}
}

func (t *Terrafile) mergeValues(parent *Terrafile) {
	tfValues := t.Values()
	if t.Values() == nil {
		tfValues = make(map[string]cty.Value)
	}
	for name, value := range parent.Values() {
		if _, ok := tfValues[name]; !ok {
			tfValues[name] = value
		}
	}
	t.ValuesBlock = &TerraValues{
		Values: tfValues,
	}
}

// mergeTemplates merges templates from a parent to a terrafile.
// Templates are matched by the name, and if a parent template does not match
// a child template, it should be added to the list (i.e. child templates
// override parent templates)
func (t *Terrafile) mergeTemplates(parent *Terrafile) {
	var mergeTemplates = make([]*TerraTemplate, 0)
	for _, parentTemplate := range parent.Templates {
		var templateMatch bool
		for _, childTemplate := range t.Templates {
			if parentTemplate.Name == childTemplate.Name {
				templateMatch = true
				break
			}
		}

		if templateMatch {
			continue
		}
		mergeTemplates = append(mergeTemplates, parentTemplate)
	}
	t.Templates = append(t.Templates, mergeTemplates...)
}

func (t *Terrafile) mergeTerraformBlock(parent *Terrafile) {
	if t.TerraformBlock == nil {
		t.TerraformBlock = &TerraformBlock{}
	}
	block := t.TerraformBlock

	if block.RequiredVersion == "" {
		block.RequiredVersion = parent.TerraformBlock.RequiredVersion
	}
	var requiredProviders = block.RequiredProviders()
	if requiredProviders == nil {
		requiredProviders = make(map[string]RequiredProvider)
	}
	for name, value := range parent.TerraformBlock.RequiredProviders() {
		if _, ok := requiredProviders[name]; !ok {
			requiredProviders[name] = value
		}
	}
	block.RequiredProvidersBlock = &TerraRequiredProviders{
		RequiredProviders: requiredProviders,
	}
}

func (t *Terrafile) Locals() map[string]cty.Value {
	if t.LocalsBlock == nil {
		return nil
	}
	return t.LocalsBlock.Locals
}

func (t *Terrafile) Variables() map[string]cty.Value {
	if t.VariablesBlock == nil {
		return nil
	}
	return t.VariablesBlock.Variables
}

func (t *Terrafile) Values() map[string]cty.Value {
	if t.ValuesBlock == nil {
		return nil
	}
	return t.ValuesBlock.Values
}

func (t *Terrafile) LocalsAsGo() (map[string]interface{}, error) {
	locals, err := fromCtyValues(t.Locals())
	if err != nil {
		return nil, fmt.Errorf("converting Cty locals{} values into Go values: %w", err)
	}
	return locals, err
}

func (t *Terrafile) VariablesAsGo() (map[string]interface{}, error) {
	vars, err := fromCtyValues(t.Variables())
	if err != nil {
		return nil, fmt.Errorf("converting Cty variables{} values into Go values: %w", err)
	}
	return vars, err
}

func (t *Terrafile) ValuesAsGo() (map[string]interface{}, error) {
	values, err := fromCtyValues(t.Values())
	if err != nil {
		return nil, fmt.Errorf("converting Cty values{} values into Go values: %w", err)
	}
	return values, err

}

func fromCtyValues(values map[string]cty.Value) (map[string]interface{}, error) {
	var retValues = make(map[string]interface{})
	for name, value := range values {
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

		retValues[name] = val
	}
	return retValues, nil
}
