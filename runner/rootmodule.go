package runner

import (
	"errors"
	"sync"

	"github.com/verifa/terraplate/parser"
)

var (
	ErrRunInProgress = errors.New("run is already in progress")
	ErrRunSkipped    = errors.New("cannot run skipped module")
)

// func newRootModule(tf *parser.Terrafile, opts TerraRunOpts) *RootModule {
func newRootModule(tf *parser.Terrafile, opts TerraRunOpts) *RootModule {
	return &RootModule{
		Terrafile: tf,
		Opts:      opts,
	}
}

type RootModule struct {
	// Terrafile is the terrafile for which this run was executed
	Terrafile *parser.Terrafile
	Opts      TerraRunOpts

	// Run stores the current run, if one has been scheduled
	Run *TerraRun
	mu  sync.RWMutex
}

// ScheduleRun schedules a run on the RootModule by setting the state and
// adding to the waitgroup.
// If a run is already in progress an error is returned and the state is unchanged
func (r *RootModule) ScheduleRun(runQueue chan *TerraRun) error {
	return r.ScheduleRunWithOpts(runQueue, r.Opts)
}

// ScheduleRun schedules a run on the RootModule by setting the state and
// adding to the waitgroup.
// If a run is already in progress an error is returned and the state is unchanged
func (r *RootModule) ScheduleRunWithOpts(runQueue chan *TerraRun, opts TerraRunOpts) error {
	// Don't schedule runs for modules that should be skipped
	if r.Skip() {
		return ErrRunSkipped
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.Run != nil && r.Run.IsRunning() {
		return ErrRunInProgress
	}
	newRun := newRunForQueue(r.Terrafile, opts)
	r.Run = newRun
	runQueue <- newRun
	return nil
}

func (r *RootModule) Wait() {
	if r.Run != nil {
		r.Run.Wait()
	}
}

func (r *RootModule) Skip() bool {
	return r.Terrafile.ExecBlock.Skip
}

func (r *RootModule) IsRunning() bool {
	if r.Run == nil {
		return false
	}
	return r.Run.IsRunning()
}

func (r *RootModule) HasRun() bool {
	return r.Run != nil
}
