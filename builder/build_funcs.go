package builder

import (
	"fmt"
	"text/template"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

func builtinFuncs() template.FuncMap {
	return template.FuncMap{
		"ctyValueToString": ctyValueToString,
	}
}

func ctyValueToString(val cty.Value) string {
	toks := hclwrite.TokensForValue(val)
	return fmt.Sprintf("%s\n", hclwrite.Format(toks.Bytes()))
}
