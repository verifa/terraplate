package runner

import (
	"errors"
	"fmt"
	"sync"

	tfjson "github.com/hashicorp/terraform-json"
	"github.com/verifa/terraplate/parser"
)

var (
	ErrRunInProgress = errors.New("run is already in progress")
)

type TerraRun struct {
	// Terrafile is the terrafile for which this run was executed
	Terrafile *parser.Terrafile

	Tasks     []*TaskResult
	Cancelled bool
	Skipped   bool

	Plan     *tfjson.Plan
	PlanText []byte

	mu        sync.RWMutex
	isRunning bool
}

// Run performs the run for this TerraRun i.e. invoking Terraform
func (r *TerraRun) Run(opts TerraRunOpts) error {
	if startErr := r.startRun(); startErr != nil {
		return startErr
	}

	tf := r.Terrafile
	// Check if root module should be skipped or not
	if tf.ExecBlock.Skip {
		fmt.Printf("%s: Skipping...\n", tf.RelativeDir())
		r.Skipped = true
		return nil
	}

	if opts.init {
		taskResult := initCmd(opts, tf)
		r.Tasks = append(r.Tasks, taskResult)
		if taskResult.HasError() {
			return nil
		}
	}
	if opts.validate {
		taskResult := validateCmd(opts, tf)
		r.Tasks = append(r.Tasks, taskResult)
		if taskResult.HasError() {
			return nil
		}
	}
	if opts.plan {
		taskResult := planCmd(opts, tf)
		r.Tasks = append(r.Tasks, taskResult)
		if taskResult.HasError() {
			return nil
		}
	}
	if opts.showPlan {
		taskResult := showPlanCmd(opts, tf)
		r.ProcessPlan(taskResult)
		r.Tasks = append(r.Tasks, taskResult)
		if taskResult.HasError() {
			return nil
		}
	}
	if opts.apply {
		taskResult := applyCmd(opts, tf)
		r.Tasks = append(r.Tasks, taskResult)
		if taskResult.HasError() {
			return nil
		}
	}

	r.endRun()
	return nil
}

func (r *TerraRun) startRun() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.isRunning {
		return ErrRunInProgress
	}
	r.isRunning = true
	return nil
}

func (r *TerraRun) endRun() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.isRunning = false
}

// PlanSummary returns a string summary to show after a plan
func (r *TerraRun) PlanSummary() string {
	// If the run had errors, we want to show that
	if r.HasError() {
		return errorColor.Sprint("Error occurred")
	}
	if r.Cancelled {
		return runCancelled.Sprint("Cancelled")
	}
	if r.Skipped {
		return "Skipped"
	}
	if !r.HasPlan() {
		return planNotAvailable.Sprint("Plan not available")
	}
	return r.Drift().Diff()
}

func (r *TerraRun) Drift() *Drift {
	if !r.HasPlan() {
		// Return an empty drift which means no drift (though user should check
		// if plan was available as well)
		return &Drift{}
	}
	return driftFromPlan(r.Plan)
}

func (r *TerraRun) HasError() bool {
	for _, task := range r.Tasks {
		if task.HasError() {
			return true
		}
	}
	return false
}

func (r *TerraRun) Errors() []error {
	var errors []error
	for _, task := range r.Tasks {
		if task.HasError() {
			errors = append(errors, task.Error)
		}
	}
	return errors
}

func (r *TerraRun) HasRelevantTasks() bool {
	for _, task := range r.Tasks {
		if task.HasRelevance() {
			return true
		}
	}
	return false
}

func (r *TerraRun) HasPlan() bool {
	return r.Plan != nil
}

// ProcessPlanText takes a TaskResult from a terraform show (without -json option)
// which makes for a compact human-readable output which we can show instead of
// the raw output from terraform plan
func (r *TerraRun) ProcessPlan(task *TaskResult) error {
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
