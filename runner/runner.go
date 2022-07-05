package runner

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/remeh/sizedwaitgroup"
	"github.com/verifa/terraplate/parser"
)

type terraCmd string

func (t terraCmd) Action() string {
	switch t {
	case terraValidate:
		return "validating"
	case terraInit:
		return "initializing"
	case terraPlan:
		return "planning"
	case terraApply:
		return "applying"
	case terraShowPlan:
		return "summarizing"
	}
	return "Unknown action"
}

const (
	terraExe               = "terraform"
	terraValidate terraCmd = "validate"
	terraInit     terraCmd = "init"
	terraPlan     terraCmd = "plan"
	terraApply    terraCmd = "apply"
	terraShowPlan terraCmd = "show"

	DefaultJobs = 4
)

func New(config *parser.TerraConfig, opts ...func(r *TerraRunOpts)) *Runner {
	// Initialise TerraRunOpts with defaults
	runOpts := TerraRunOpts{
		jobs: DefaultJobs,
	}
	for _, opt := range opts {
		opt(&runOpts)
	}

	runner := Runner{
		ctx:    listenTerminateSignals(),
		swg:    sizedwaitgroup.New(runOpts.jobs),
		opts:   runOpts,
		config: config,
	}

	// Initialize result
	var (
		rootMods = config.RootModules()
		runs     = make([]*TerraRun, len(rootMods))
	)
	for index, tf := range config.RootModules() {
		runs[index] = &TerraRun{
			Terrafile: tf,
		}
	}
	runner.Runs = runs
	return &runner
}

type Runner struct {
	opts TerraRunOpts
	ctx  context.Context
	swg  sizedwaitgroup.SizedWaitGroup

	config *parser.TerraConfig

	Runs []*TerraRun
}

func Run(config *parser.TerraConfig, opts ...func(r *TerraRunOpts)) *Runner {

	runner := New(config, opts...)
	runner.RunAll()
	return runner
}

func Jobs(jobs int) func(r *TerraRunOpts) {
	return func(r *TerraRunOpts) {
		r.jobs = jobs
	}
}

func RunValidate() func(r *TerraRunOpts) {
	return func(r *TerraRunOpts) {
		r.validate = true
	}
}

func RunInit() func(r *TerraRunOpts) {
	return func(r *TerraRunOpts) {
		r.init = true
	}
}

func RunPlan() func(r *TerraRunOpts) {
	return func(r *TerraRunOpts) {
		r.plan = true
	}
}

func RunShowPlan() func(r *TerraRunOpts) {
	return func(r *TerraRunOpts) {
		r.showPlan = true
	}
}

func RunApply() func(r *TerraRunOpts) {
	return func(r *TerraRunOpts) {
		r.apply = true
	}
}

func ExtraArgs(extraArgs []string) func(r *TerraRunOpts) {
	return func(r *TerraRunOpts) {
		r.extraArgs = extraArgs
	}
}

// TerraRunOpts handles running Terraform over the root modules
type TerraRunOpts struct {
	validate bool
	init     bool
	plan     bool
	showPlan bool
	apply    bool

	// Max number of concurrent jobs allowed
	jobs int
	// Terraform command flags
	extraArgs []string
}

func (r *Runner) RunAll() {
	for _, run := range r.Runs {
		addErr := r.swg.AddWithContext(r.ctx)
		// Check if the process has been cancelled.
		if errors.Is(addErr, context.Canceled) {
			run.Cancelled = true
			continue
		}

		// Set local run for goroutine
		run := run
		go func() {
			defer r.swg.Done()
			run.Run(r.opts)
		}()
	}
	r.swg.Wait()
}

// Log returns a string of the runs and tasks to print to the console
func (r *Runner) Log() string {
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
func (r *Runner) PlanSummary() string {
	var summary strings.Builder
	summary.WriteString(boldColor.Sprint("\nTerraplate Plan Summary\n\n"))
	for _, run := range r.Runs {
		summary.WriteString(fmt.Sprintf("%s: %s\n", run.Terrafile.RelativeDir(), run.PlanSummary()))
	}
	return summary.String()
}

func (r *Runner) RunsWithDrift() []*TerraRun {
	var runs []*TerraRun
	for _, run := range r.Runs {
		if run.Drift().HasDrift() {
			runs = append(runs, run)
		}
	}
	return runs
}

func (r *Runner) RunsWithError() []*TerraRun {
	var runs []*TerraRun
	for _, run := range r.Runs {
		if run.HasError() {
			runs = append(runs, run)
		}
	}
	return runs
}

// HasDrift returns true if any drift was detected in any of the runs
func (r *Runner) HasDrift() bool {
	for _, run := range r.Runs {
		if drift := run.Drift(); drift != nil {
			// If at least one of the runs has drifted, then our result has drift
			if drift.HasDrift() {
				return true
			}
		}
	}
	return false
}

func (r *Runner) HasError() bool {
	for _, run := range r.Runs {
		if run.HasError() {
			return true
		}
	}
	return false
}

// Errors returns a multierror with any errors found in any tasks within the runs
func (r *Runner) Errors() error {
	var err error
	for _, run := range r.Runs {
		if run.HasError() {
			err = multierror.Append(err, run.Errors()...)
		}
	}
	return err
}
