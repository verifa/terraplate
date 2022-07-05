package notify

import (
	"bytes"
	"fmt"

	texttemplate "text/template"

	"github.com/verifa/terraplate/runner"
)

var (
	statusErrorColor = "#d2222d"
	statusDrftColor  = "#ffbf00"
	statusSyncColor  = "#238823"
)

type Data struct {
	Runner     *runner.Runner
	Repo       *Repo
	ResultsURL string
}

func (d Data) StatusColor() string {
	switch {
	case d.Runner.HasError():
		return statusErrorColor
	case d.Runner.HasDrift():
		return statusDrftColor
	}
	return statusSyncColor
}

func execTemplate(text string, data *Data) ([]byte, error) {
	tmpl, tmplErr := texttemplate.New("notification").
		Option("missingkey=error").
		Parse(text)
	if tmplErr != nil {
		return nil, tmplErr
	}
	var rawContents bytes.Buffer
	if execErr := tmpl.Execute(&rawContents, data); execErr != nil {
		return nil, fmt.Errorf("executing template: %w", execErr)
	}
	return rawContents.Bytes(), nil
}
