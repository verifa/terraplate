package modulesui

import (
	"fmt"
	"strings"

	"github.com/verifa/terraplate/runner"
)

type listItem struct {
	module *runner.RootModule
}

func (i listItem) Title() string {
	return i.module.Terrafile.Dir
}

func (i listItem) Description() string {
	return i.renderSummary()
}

func (i listItem) FilterValue() string {
	return i.module.Terrafile.Dir
}

func (i listItem) renderSummary() string {
	var (
		summary strings.Builder
		module  = i.module
		run     = module.Run
		drift   *runner.Drift
	)
	if run != nil {
		drift = run.Drift()
	}
	switch {
	case module.Skip():
		summary.WriteString("Skip.")
	case run == nil:
		summary.WriteString("Not run.")
	case run.IsRunning():
		summary.WriteString("Running...")
	case run.HasError():
		summary.WriteString("Error occurred.")
	case run.IsApplied():
		summary.WriteString("Applied.")
	case !run.HasPlan():
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
