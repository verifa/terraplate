package runner

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

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

func Run(config *parser.TerraConfig, opts ...func(r *TerraRun)) *Result {
	// Initialise TerraRun with defaults
	run := TerraRun{
		jobs: DefaultJobs,
	}
	for _, opt := range opts {
		opt(&run)
	}

	// Listen to terminate calls (i.e. SIGINT) instead of exiting immediately
	ctx := listenTerminateSignals()

	var (
		rootMods   = config.RootModules()
		runResults = make([]*RunResult, len(rootMods))
	)
	// Start terraform runs in different root modules based on number of concurrent
	// jobs that are allowed
	swg := sizedwaitgroup.New(run.jobs)
	for index, tf := range config.RootModules() {

		addErr := swg.AddWithContext(ctx)
		// Check if the process has been cancelled.
		if errors.Is(addErr, context.Canceled) {
			// Set an empty RunResult
			runResults[index] = &RunResult{
				Terrafile: tf,
				Cancelled: true,
			}
			continue
		}
		tf := tf
		index := index

		go func() {
			defer swg.Done()
			result := runCmds(&run, tf)
			runResults[index] = result
		}()

	}
	swg.Wait()

	return &Result{
		Runs: runResults,
	}
}

func Jobs(jobs int) func(r *TerraRun) {
	return func(r *TerraRun) {
		r.jobs = jobs
	}
}

func RunValidate() func(r *TerraRun) {
	return func(r *TerraRun) {
		r.validate = true
	}
}

func RunInit() func(r *TerraRun) {
	return func(r *TerraRun) {
		r.init = true
	}
}

func RunPlan() func(r *TerraRun) {
	return func(r *TerraRun) {
		r.plan = true
	}
}

func RunShowPlan() func(r *TerraRun) {
	return func(r *TerraRun) {
		r.showPlan = true
	}
}

func RunApply() func(r *TerraRun) {
	return func(r *TerraRun) {
		r.apply = true
	}
}

func ExtraArgs(extraArgs []string) func(r *TerraRun) {
	return func(r *TerraRun) {
		r.extraArgs = extraArgs
	}
}

type TerraRun struct {
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

func runCmds(run *TerraRun, tf *parser.Terrafile) *RunResult {
	result := RunResult{
		Terrafile: tf,
	}
	// Check if root module should be skipped or not
	if tf.ExecBlock.Skip {
		fmt.Printf("%s: Skipping...\n", tf.RelativeDir())
		result.Skipped = true
		return &result
	}

	if run.init {
		taskResult := initCmd(run, tf)
		result.Tasks = append(result.Tasks, taskResult)
		if taskResult.HasError() {
			return &result
		}
	}
	if run.validate {
		taskResult := validateCmd(run, tf)
		result.Tasks = append(result.Tasks, taskResult)
		if taskResult.HasError() {
			return &result
		}
	}
	if run.plan {
		taskResult := planCmd(run, tf)
		result.Tasks = append(result.Tasks, taskResult)
		if taskResult.HasError() {
			return &result
		}
	}
	if run.showPlan {
		taskResult := showPlanCmd(run, tf)
		result.ProcessPlan(taskResult)
		result.Tasks = append(result.Tasks, taskResult)
		if taskResult.HasError() {
			return &result
		}
	}
	if run.apply {
		taskResult := applyCmd(run, tf)
		result.Tasks = append(result.Tasks, taskResult)
		if taskResult.HasError() {
			return &result
		}
	}
	return &result
}

func initCmd(run *TerraRun, tf *parser.Terrafile) *TaskResult {
	var args []string
	args = append(args, tfCleanExtraArgs(run.extraArgs)...)
	return runCmd(tf, terraInit, args)
}

func validateCmd(run *TerraRun, tf *parser.Terrafile) *TaskResult {
	var args []string
	args = append(args, tfCleanExtraArgs(run.extraArgs)...)
	return runCmd(tf, terraValidate, args)
}

func planCmd(run *TerraRun, tf *parser.Terrafile) *TaskResult {
	plan := tf.ExecBlock.PlanBlock

	var args []string
	args = append(args,
		fmt.Sprintf("-lock=%v", plan.Lock),
		fmt.Sprintf("-input=%v", plan.Input),
	)
	if !plan.SkipOut {
		args = append(args,
			"-out="+plan.Out,
		)
	}
	args = append(args, tfCleanExtraArgs(tf.ExecBlock.ExtraArgs)...)
	args = append(args, tfCleanExtraArgs(run.extraArgs)...)
	return runCmd(tf, terraPlan, args)
}

func showPlanCmd(run *TerraRun, tf *parser.Terrafile) *TaskResult {
	plan := tf.ExecBlock.PlanBlock
	if plan.SkipOut {
		return &TaskResult{
			TerraCmd: terraShowPlan,
			Skipped:  true,
		}
	}
	var args []string
	args = append(args, "-json", plan.Out)
	return runCmd(tf, terraShowPlan, args)
}

func applyCmd(run *TerraRun, tf *parser.Terrafile) *TaskResult {
	plan := tf.ExecBlock.PlanBlock

	var args []string
	args = append(args,
		fmt.Sprintf("-lock=%v", plan.Lock),
		fmt.Sprintf("-input=%v", plan.Input),
	)
	args = append(args, tfCleanExtraArgs(tf.ExecBlock.ExtraArgs)...)
	args = append(args, tfCleanExtraArgs(run.extraArgs)...)

	if !plan.SkipOut {
		args = append(args, plan.Out)
	}

	return runCmd(tf, terraApply, args)
}

func runCmd(tf *parser.Terrafile, tfCmd terraCmd, args []string) *TaskResult {
	result := TaskResult{
		TerraCmd: tfCmd,
	}
	cmdArgs := append(tfArgs(tf), string(tfCmd))
	cmdArgs = append(cmdArgs, args...)
	result.ExecCmd = exec.Command(terraExe, cmdArgs...)

	// Create channel and start progress printer
	done := make(chan bool)
	go printProgress(tf.RelativeDir(), tfCmd, done)
	defer func() { done <- true }()

	var runErr error
	result.Output, runErr = result.ExecCmd.CombinedOutput()
	if runErr != nil {
		result.Error = fmt.Errorf("%s: running %s command", tf.RelativeDir(), tfCmd)
	}

	return &result
}

func tfArgs(tf *parser.Terrafile) []string {
	var args []string
	args = append(args, "-chdir="+tf.Dir)
	return args
}

// tfCleanExtraArgs returns the provided slice with any empty spaces removed.
// Empty spaces create weird errors that are hard to debug
func tfCleanExtraArgs(args []string) []string {
	var cleanArgs = make([]string, 0)
	for _, arg := range args {
		if arg != "" {
			cleanArgs = append(cleanArgs, arg)
		}
	}
	return cleanArgs
}

// listenTerminateSignals returns a context that will be cancelled if an interrupt
// or termination signal is received. The context can be used to prevent further
// runs from being scheduled
func listenTerminateSignals() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		for {
			<-signals
			fmt.Println("")
			fmt.Println("Terraplate: Interrupt received.")
			fmt.Println("Terraplate: Sending interrupt to all Terraform processes and cancelling any queued runs.")
			// Cancel the context, to stop any more runs from being executed
			cancel()
		}
	}()
	return ctx
}

func printProgress(path string, cmd terraCmd, done <-chan bool) {
	var (
		interval = time.Second * 10
		ticker   = time.NewTicker(interval)
		elapsed  time.Duration
	)
	defer ticker.Stop()
	// Print initial line
	fmt.Printf("%s: %s...\n", path, strings.Title(cmd.Action()))
	for {
		select {
		case <-ticker.C:
			elapsed += interval
			fmt.Printf("%s: Still %s... [%s elapsed]\n", path, cmd.Action(), elapsed)
		case <-done:
			return
		}
	}
}
