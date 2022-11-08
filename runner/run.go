package runner

import (
	"fmt"
	"strings"
	"sync"

	tfjson "github.com/hashicorp/terraform-json"
	"github.com/verifa/terraplate/parser"
)

type state int

const (
	finishedState state = iota
	queueState
	runState
)

func newRunForQueue(tf *parser.Terrafile, opts TerraRunOpts) *TerraRun {

	var r TerraRun
	r.Terrafile = tf
	r.Opts = opts
	r.state = queueState
	// Increment waitgroup so that we can wait on this run, even whilst it is
	// queueing, until it is finished
	r.wg.Add(1)
	return &r
}

type TerraRun struct {
	// Terrafile is the terrafile for which this run was executed
	Terrafile *parser.Terrafile
	Opts      TerraRunOpts

	Tasks     []*TaskResult
	Cancelled bool

	Plan     *tfjson.Plan
	PlanText []byte

	wg    sync.WaitGroup
	state state
}

// Run performs a blocking run for this TerraRun i.e. invoking Terraform
func (r *TerraRun) Run() {
	r.Start()
	r.Wait()
}

// Start performs a non-blocking run for this TerraRun i.e. invoking Terraform
func (r *TerraRun) Start() {
	r.startRun()
	defer r.endRun()

	tf := r.Terrafile

	if r.Opts.build {
		taskResult := buildCmd(r.Opts, tf)
		r.Tasks = append(r.Tasks, taskResult)
		if taskResult.HasError() {
			return
		}
	}
	if r.Opts.init {
		taskResult := initCmd(r.Opts, tf)
		r.Tasks = append(r.Tasks, taskResult)
		if taskResult.HasError() {
			return
		}
	}
	if r.Opts.validate {
		taskResult := validateCmd(r.Opts, tf)
		r.Tasks = append(r.Tasks, taskResult)
		if taskResult.HasError() {
			return
		}
	}
	if r.Opts.plan {
		taskResult := planCmd(r.Opts, tf)
		r.Tasks = append(r.Tasks, taskResult)
		if taskResult.HasError() {
			return
		}
	}
	if r.Opts.show {
		taskResult := showCmd(r.Opts, tf)
		r.Tasks = append(r.Tasks, taskResult)
		if taskResult.HasError() {
			return
		}
	}
	if r.Opts.showPlan {
		taskResult := showPlanCmd(r.Opts, tf)
		r.ProcessPlan(taskResult)
		r.Tasks = append(r.Tasks, taskResult)
		if taskResult.HasError() {
			return
		}
	}
	if r.Opts.apply {
		taskResult := applyCmd(r.Opts, tf)
		r.Tasks = append(r.Tasks, taskResult)
		if taskResult.HasError() {
			return
		}
	}
}

// Wait blocks and waits for the run to be finished
func (r *TerraRun) Wait() {
	r.wg.Wait()
}

func (r *TerraRun) startRun() {
	r.state = runState
}

func (r *TerraRun) endRun() {
	defer r.wg.Done()
	r.state = finishedState
}

func (r *TerraRun) Log(fullLog bool) string {
	var log strings.Builder
	log.WriteString(boldColor.Sprintf("Run for %s\n\n", r.Terrafile.Dir))

	for _, task := range r.Tasks {
		if fullLog || task.IsRelevant() {
			log.WriteString(task.Log())
		}
	}
	return log.String()
}

// Summary returns a string summary to show after a plan
func (r *TerraRun) Summary() string {
	switch {
	case r.HasError():
		return errorColor.Sprint("Error occurred")
	case r.Cancelled:
		return runCancelled.Sprint("Cancelled")
	case r.IsApplied():
		return boldColor.Sprint("Applied")
	case r.IsPlanned():
		if !r.HasPlan() {
			return planNotAvailable.Sprint("Plan not available")
		}
		return r.Drift().Diff()
	case r.IsInitd():
		return boldColor.Sprint("Initialized")
	case r.IsBuilt():
		return boldColor.Sprint("Built")
	default:
		return "Unknown status"
	}
}

func (r *TerraRun) Drift() *Drift {
	if r == nil || !r.HasPlan() {
		// Return an empty drift which means no drift (though user should check
		// if plan was available as well)
		return &Drift{}
	}
	return driftFromPlan(r.Plan)
}

func (r *TerraRun) IsApplied() bool {
	if r == nil {
		return false
	}
	if r.state != finishedState {
		return false
	}
	for _, task := range r.Tasks {
		if task.TerraCmd == terraApply {
			return true
		}
	}
	return false
}

func (r *TerraRun) IsPlanned() bool {
	if r == nil {
		return false
	}
	if r.state != finishedState {
		return false
	}
	for _, task := range r.Tasks {
		switch task.TerraCmd {
		case terraPlan, terraShowJSON:
			return true
		}
	}
	return false
}

func (r *TerraRun) IsInitd() bool {
	if r == nil {
		return false
	}
	if r.state != finishedState {
		return false
	}
	for _, task := range r.Tasks {
		if task.TerraCmd == terraInit {
			return true
		}
	}
	return false
}

func (r *TerraRun) IsBuilt() bool {
	if r == nil {
		return false
	}
	if r.state != finishedState {
		return false
	}
	for _, task := range r.Tasks {
		if task.TerraCmd == terraBuild {
			return true
		}
	}
	return false
}

func (r *TerraRun) IsRunning() bool {
	if r == nil {
		return false
	}
	return r.state == queueState || r.state == runState
}

func (r *TerraRun) HasError() bool {
	if r == nil {
		return false
	}
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
		if task.IsRelevant() {
			return true
		}
	}
	return false
}

func (r *TerraRun) HasPlan() bool {
	if r == nil {
		return false
	}
	return r.Plan != nil
}

// ProcessPlanText takes a TaskResult from a terraform show (with -json option)
// which makes for a compact human-readable output which we can show instead of
// the raw output from terraform plan
func (r *TerraRun) ProcessPlan(task *TaskResult) error {
	// Make sure we received a `terraform show` task result
	if task.TerraCmd != terraShowJSON {
		return fmt.Errorf("terraform show command required for processing plan: received %s", task.TerraCmd)
	}
	// Cannot process a plan if the `terraform show` command error'd
	if task.HasError() {
		return nil
	}
	var tfPlan tfjson.Plan
	if err := tfPlan.UnmarshalJSON(task.Output.Bytes()); err != nil {
		return fmt.Errorf("unmarshalling terraform show plan output: %w", err)
	}

	r.Plan = &tfPlan

	return nil
}
