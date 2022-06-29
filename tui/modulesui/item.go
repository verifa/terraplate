package modulesui

import (
	"fmt"
	"strings"

	"github.com/verifa/terraplate/runner"
)

type listItem struct {
	run *runner.RunResult
}

func (i listItem) Title() string {
	return i.run.Terrafile.Dir
}

func (i listItem) Description() string {

	return i.renderSummary()
}

func (i listItem) FilterValue() string {
	return i.run.Terrafile.Dir
}

func (i listItem) renderSummary() string {
	var (
		summary strings.Builder
		drift   = i.run.Drift()
	)
	switch {
	case i.run.HasError():
		summary.WriteString("Error occurred.")
	case !i.run.HasPlan():
		summary.WriteString("No plan available.")
	case !drift.HasDrift():
		summary.WriteString("No changes.")
	default:
		// There is a plan, and there is drift, so show a summary
		fmt.Fprintf(
			&summary,
			"+ %d   ~ %d   - %d",
			len(drift.AddResources),
			len(drift.ChangeResources),
			len(drift.DestroyResources),
		)
	}
	return summary.String()
}
