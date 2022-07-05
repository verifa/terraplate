package runner

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/verifa/terraplate/parser"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func initCmd(run TerraRunOpts, tf *parser.Terrafile) *TaskResult {
	var args []string
	args = append(args, tfCleanExtraArgs(run.extraArgs)...)
	return runCmd(tf, terraInit, args)
}

func validateCmd(run TerraRunOpts, tf *parser.Terrafile) *TaskResult {
	var args []string
	args = append(args, tfCleanExtraArgs(run.extraArgs)...)
	return runCmd(tf, terraValidate, args)
}

func planCmd(run TerraRunOpts, tf *parser.Terrafile) *TaskResult {
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

func showPlanCmd(run TerraRunOpts, tf *parser.Terrafile) *TaskResult {
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

func applyCmd(run TerraRunOpts, tf *parser.Terrafile) *TaskResult {
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
	fmt.Printf("%s: %s...\n", path, cases.Title(language.English).String(cmd.Action()))
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
