package runner

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/hashicorp/go-multierror"
	"github.com/verifa/terraplate/parser"
)

const (
	DefaultJobs = 4
	maxRunQueue = 200
)

func Run(config *parser.TerraConfig, opts ...func(r *TerraRunOpts)) *Runner {
	runner := New(config, opts...)
	runner.RunAll()
	return runner
}

func New(config *parser.TerraConfig, opts ...func(r *TerraRunOpts)) *Runner {
	runOpts := NewOpts(opts...)

	// Create the runQueue which is used by the workers to schedule runs
	runQueue := make(chan *TerraRun, maxRunQueue)

	runner := Runner{
		ctx:      listenTerminateSignals(runQueue),
		runQueue: runQueue,
		config:   config,
		Opts:     runOpts,
	}
	// Initialize the workers in separate go routines
	for workerID := 0; workerID < runOpts.jobs; workerID++ {
		go runner.startWorker(workerID)
	}

	// Initialize result
	var (
		// Get only root module Terrafiles
		tfs     = config.RootModules()
		modules = make([]*RootModule, len(tfs))
	)
	for index, tf := range tfs {
		modules[index] = newRootModule(tf, runOpts)
	}
	runner.Modules = modules
	return &runner
}

type Runner struct {
	ctx context.Context
	// runQueue is a channel for managing the run queue
	runQueue chan *TerraRun
	wg       sync.WaitGroup

	config *parser.TerraConfig

	Opts    TerraRunOpts
	Modules []*RootModule
}

func (r *Runner) WorkingDirectory() string {
	return r.config.WorkingDirectory
}

func (r *Runner) RunAll() {
	r.Run(r.Modules)
}

func (r *Runner) Run(modules []*RootModule) {
	r.Start(modules)
	r.Wait()
}

func (r *Runner) Start(modules []*RootModule) {
	r.StartWithOpts(modules, r.Opts)
}

func (r *Runner) StartWithOpts(modules []*RootModule, opts TerraRunOpts) {
	for _, mod := range modules {
		// Check that the run is not in progress
		if runErr := mod.ScheduleRunWithOpts(r.runQueue, opts); runErr != nil {
			continue
		}
		// If run was scheduled, add to waitgroup
		r.wg.Add(1)
	}
}

func (r *Runner) Wait() {
	r.wg.Wait()
}

// Runs returns the list of latest runs for the root modules
func (r *Runner) Runs() []*TerraRun {
	var runs = make([]*TerraRun, 0)
	for _, mod := range r.Modules {
		if mod.Run != nil {
			runs = append(runs, mod.Run)
		}
	}
	return runs
}

// Log returns a string of the runs and tasks to print to the console
func (r *Runner) Log(level OutputLevel) string {
	var (
		summary         strings.Builder
		hasRelevantRuns bool
	)
	summary.WriteString(textSeparator)
	for _, run := range r.Runs() {
		// Skip runs that have nothing relevant to show
		if !run.HasRelevantTasks(level) {
			continue
		}
		hasRelevantRuns = true

		summary.WriteString(run.Log(false))
	}
	// If there were no runs to output, return an empty string to avoid printing
	// separators and empty space
	if !hasRelevantRuns {
		return ""
	}
	summary.WriteString(textSeparator)
	return summary.String()
}

// Summary returns a string summary to show after a plan
func (r *Runner) Summary(level OutputLevel) string {
	var (
		summary         strings.Builder
		hasRelevantRuns bool
	)
	summary.WriteString(boldColor.Sprint("\nTerraplate Summary\n\n"))
	for _, run := range r.Runs() {
		showSummary := level.ShowAll() || (run.Drift().HasDrift() && level.ShowDrift()) || run.HasError()

		if showSummary {
			hasRelevantRuns = true
			summary.WriteString(fmt.Sprintf("%s: %s\n", run.Terrafile.Dir, run.Summary()))
		}
	}
	if !hasRelevantRuns {
		summary.WriteString("Everything up to date: no drift and no errors\n")
	}
	return summary.String()
}

func (r *Runner) RunsWithDrift() []*TerraRun {
	var runs []*TerraRun
	for _, run := range r.Runs() {
		if run.Drift().HasDrift() {
			runs = append(runs, run)
		}
	}
	return runs
}

func (r *Runner) RunsWithError() []*TerraRun {
	var runs []*TerraRun
	for _, run := range r.Runs() {
		if run.HasError() {
			runs = append(runs, run)
		}
	}
	return runs
}

// HasDrift returns true if any drift was detected in any of the runs
func (r *Runner) HasDrift() bool {
	for _, run := range r.Runs() {
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
	for _, run := range r.Runs() {
		if run.HasError() {
			return true
		}
	}
	return false
}

// Errors returns a multierror with any errors found in any tasks within the runs
func (r *Runner) Errors() error {
	var err error
	for _, run := range r.Runs() {
		if run.HasError() {
			err = multierror.Append(err, run.Errors()...)
		}
	}
	return err
}

// listenTerminateSignals returns a context that will be cancelled if an interrupt
// or termination signal is received. The context can be used to prevent further
// runs from being scheduled
func listenTerminateSignals(runQueue chan *TerraRun) context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		for {
			<-signals
			fmt.Println("")
			fmt.Println("Terraplate: Interrupt received.")
			fmt.Println("Terraplate: Sending interrupt to all Terraform processes and cancelling any queued runs.")
			fmt.Println("")
			// Cancel the context and stop any more runs from being executed
			cancel()
			close(runQueue)
		}
	}()
	return ctx
}
