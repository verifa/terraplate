package parser

// BuildData defines the data which is passed to the Go template engine
type BuildData struct {
	Locals    map[string]interface{}
	Variables map[string]interface{}
	Values    map[string]interface{}
	Terrafile *Terrafile
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
