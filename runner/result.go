package runner

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/fatih/color"
	"github.com/hashicorp/go-multierror"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/verifa/terraplate/parser"
)

var (
	boldColor          = color.New(color.Bold)
	errorColor         = color.New(color.FgRed, color.Bold)
	planNotAvailable   = color.New(color.FgMagenta, color.Bold)
	planNoChangesColor = color.New(color.FgGreen, color.Bold)
	planCreateColor    = color.New(color.FgGreen, color.Bold)
	planDestroyColor   = color.New(color.FgRed, color.Bold)
	planUpdateColor    = color.New(color.FgYellow, color.Bold)
)

var (
	textSeparator = boldColor.Sprint("\n─────────────────────────────────────────────────────────────────────────────\n\n")
)

type Result struct {
	Runs []*RunResult
}

// Log returns a string of the runs and tasks to print to the console
func (r *Result) Log() string {
	var (
		summary         strings.Builder
		hasRelevantRuns bool
	)
	summary.WriteString(textSeparator)
	for _, run := range r.Runs {
		// Skip runs that have nothing relevant to show
		if !run.HasRelevantTasks() {
			continue
		}
		hasRelevantRuns = true

		summary.WriteString(boldColor.Sprintf("Run for %s\n\n", run.Terrafile.RelativeDir()))

		for _, task := range run.Tasks {
			if task.HasRelevance() {
				summary.WriteString(task.Log())
			}
		}
	}
	// If there were no runs to output, return an empty string to avoid printing
	// separators and empty space
	if !hasRelevantRuns {
		return ""
	}
	summary.WriteString(textSeparator)
	return summary.String()
}

// PlanSummary returns a string summary to show after a plan
func (r *Result) PlanSummary() string {
	var summary strings.Builder
	summary.WriteString(boldColor.Sprint("\nTerraplate Plan Summary\n\n"))
	for _, run := range r.Runs {
		summary.WriteString(fmt.Sprintf("%s: %s\n", run.Terrafile.RelativeDir(), run.PlanSummary()))
	}
	return summary.String()
}

func (r *Result) HasError() bool {
	for _, run := range r.Runs {
		if run.HasError() {
			return true
		}
	}
	return false
}

// Errors returns a multierror with any errors found in any tasks within the runs
func (r *Result) Errors() error {
	var err error
	for _, run := range r.Runs {
		if run.HasError() {
			err = multierror.Append(err, run.Errors()...)
		}
	}
	return err
}

type RunResult struct {
	// Terrafile is the terrafile for which this run was executed
	Terrafile *parser.Terrafile

	Tasks   []*TaskResult
	Skipped bool

	Plan     *tfjson.Plan
	PlanText []byte
}

// PlanSummary returns a string summary to show after a plan
func (r *RunResult) PlanSummary() string {
	// If the run had errors, we want to show that
	if r.HasError() {
		return errorColor.Sprint("Error occurred")
	}
	if r.Skipped {
		return "Skipped"
	}
	if !r.HasPlan() {
		return planNotAvailable.Sprint("Plan not available")
	}

	return tfPlanSummary(r.Plan)
}

func (r *RunResult) HasError() bool {
	for _, task := range r.Tasks {
		if task.HasError() {
			return true
		}
	}
	return false
}

func (r *RunResult) Errors() []error {
	var errors []error
	for _, task := range r.Tasks {
		if task.HasError() {
			errors = append(errors, task.Error)
		}
	}
	return errors
}

func (r *RunResult) HasRelevantTasks() bool {
	for _, task := range r.Tasks {
		if task.HasRelevance() {
			return true
		}
	}
	return false
}

func (r *RunResult) HasPlan() bool {
	return r.Plan != nil
}

// ProcessPlanText takes a TaskResult from a terraform show (without -json option)
// which makes for a compact human-readable output which we can show instead of
// the raw output from terraform plan
func (r *RunResult) ProcessPlan(task *TaskResult) error {
	// Make sure we received a `terraform show` task result
	if task.TerraCmd != terraShowPlan {
		return fmt.Errorf("terraform show command required for processing plan: received %s", task.TerraCmd)
	}
	// Cannot process a plan if the `terraform show` command error'd
	if task.HasError() {
		return nil
	}
	var tfPlan tfjson.Plan
	if err := tfPlan.UnmarshalJSON(task.Output); err != nil {
		return fmt.Errorf("unmarshalling terraform show plan output: %w", err)
	}

	r.Plan = &tfPlan

	return nil
}

type TaskResult struct {
	ExecCmd  *exec.Cmd
	TerraCmd terraCmd

	Output  []byte
	Error   error
	Skipped bool
}

func (t *TaskResult) HasError() bool {
	return t.Error != nil
}

// HasRelevance is an attempt at better UX.
// We don't simply want to output everything. Things like successful inits and
// terraform show output are not interesting for the user, so skip them by
// default and therefore keep the output less
func (t *TaskResult) HasRelevance() bool {
	// Errors are always relevant
	if t.HasError() {
		return true
	}
	// Skipped tasks are not relevant
	if t.Skipped {
		return false
	}
	switch t.TerraCmd {
	case terraPlan:
		// Plan outputs are interesting
		return true
	case terraApply:
		// Apply outputs are interesting
		return true
	default:
		// Skip other command outputs
		return false
	}
}

func (t *TaskResult) Log() string {
	var summary strings.Builder

	summary.WriteString(fmt.Sprintf("%s output: %s\n\n", strings.Title(string(t.TerraCmd)), t.ExecCmd.String()))
	if t.HasError() {
		summary.WriteString(fmt.Sprintf("Error: %s\n\n", t.Error.Error()))
	}
	scanner := bufio.NewScanner(bytes.NewBuffer(t.Output))
	for scanner.Scan() {
		summary.WriteString(fmt.Sprintf("    %s\n", scanner.Text()))
	}
	summary.WriteString("\n\n")

	return summary.String()
}

func tfPlanSummary(plan *tfjson.Plan) string {
	var (
		numCreate  = 0
		numDestroy = 0
		numChange  = 0
	)
	for _, change := range plan.ResourceChanges {
		for _, action := range change.Change.Actions {
			switch action {
			case tfjson.ActionCreate:
				numCreate++
			case tfjson.ActionDelete:
				numDestroy++
			case tfjson.ActionUpdate:
				numChange++
			default:
				// We don't care about other actions for the summary
			}

		}
	}
	if numCreate == 0 && numDestroy == 0 && numChange == 0 {
		return planNoChangesColor.Sprint("No changes.")
	}
	return fmt.Sprintf(
		"%s %s %s",
		planCreateColor.Sprintf("%d to add.", numCreate),
		planUpdateColor.Sprintf("%d to change.", numChange),
		planDestroyColor.Sprintf("%d to destroy.", numDestroy),
	)
}
