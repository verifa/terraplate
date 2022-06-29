package runner

import (
	"fmt"
	"io"
	"log"
	"os/exec"
	"time"

	"github.com/verifa/terraplate/builder"
	"github.com/verifa/terraplate/parser"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type terraCmd string

func (t terraCmd) Action() string {
	switch t {
	case terraBuild:
		return "building"
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
	terraExe = "terraform"

	terraBuild    terraCmd = "build"
	terraValidate terraCmd = "validate"
	terraInit     terraCmd = "init"
	terraPlan     terraCmd = "plan"
	terraApply    terraCmd = "apply"
	terraShowPlan terraCmd = "show"
)

func buildCmd(opts TerraRunOpts, tf *parser.Terrafile) *TaskResult {
	var task TaskResult
	task.TerraCmd = terraBuild
	task.Error = builder.BuildTerrafile(tf, &task.Output)
	return &task
}

func validateCmd(opts TerraRunOpts, tf *parser.Terrafile) *TaskResult {
	var args []string
	args = append(args, tfCleanExtraArgs(opts.extraArgs)...)
	return runCmd(opts.out, tf, terraValidate, args)
}

func initCmd(opts TerraRunOpts, tf *parser.Terrafile) *TaskResult {
	var args []string
	if opts.initUpgrade {
		args = append(args, "-upgrade")
	}
	args = append(args, tfCleanExtraArgs(opts.extraArgs)...)
	return runCmd(opts.out, tf, terraInit, args)
}

func planCmd(opts TerraRunOpts, tf *parser.Terrafile) *TaskResult {
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
	args = append(args, tfCleanExtraArgs(opts.extraArgs)...)
	return runCmd(opts.out, tf, terraPlan, args)
}

func showPlanCmd(opts TerraRunOpts, tf *parser.Terrafile) *TaskResult {
	plan := tf.ExecBlock.PlanBlock
	if plan.SkipOut {
		return &TaskResult{
			TerraCmd: terraShowPlan,
			Skipped:  true,
		}
	}
	var args []string
	args = append(args, "-json", plan.Out)
	return runCmd(opts.out, tf, terraShowPlan, args)
}

func applyCmd(opts TerraRunOpts, tf *parser.Terrafile) *TaskResult {
	plan := tf.ExecBlock.PlanBlock

	var args []string
	args = append(args,
		fmt.Sprintf("-lock=%v", plan.Lock),
		fmt.Sprintf("-input=%v", plan.Input),
	)
	args = append(args, tfCleanExtraArgs(tf.ExecBlock.ExtraArgs)...)
	args = append(args, tfCleanExtraArgs(opts.extraArgs)...)

	if !plan.SkipOut {
		args = append(args, plan.Out)
	}

	return runCmd(opts.out, tf, terraApply, args)
}

func runCmd(out io.Writer, tf *parser.Terrafile, tfCmd terraCmd, args []string) *TaskResult {
	task := TaskResult{
		TerraCmd: tfCmd,
	}
	cmdArgs := append(tfArgs(tf), string(tfCmd))
	cmdArgs = append(cmdArgs, args...)
	task.ExecCmd = exec.Command(terraExe, cmdArgs...)

	// Create channel and start progress printer
	done := make(chan bool)
	go printProgress(out, tf.Dir, tfCmd, done)
	defer func() { done <- true }()

	pr, pw := io.Pipe()
	defer pw.Close()
	task.ExecCmd.Stdout = pw
	task.ExecCmd.Stderr = pw

	if err := task.ExecCmd.Start(); err != nil {
		task.Error = fmt.Errorf("starting command: %w", err)
		return &task
	}
	go func() {
		if _, err := io.Copy(&task.Output, pr); err != nil {
			log.Fatal(err)
		}
	}()

	runErr := task.ExecCmd.Wait()
	if runErr != nil {
		task.Error = fmt.Errorf("%s: running %s command", tf.Dir, tfCmd)
	}

	return &task
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

func printProgress(out io.Writer, path string, cmd terraCmd, done <-chan bool) {
	var (
		interval = time.Second * 10
		ticker   = time.NewTicker(interval)
		elapsed  time.Duration
	)
	defer ticker.Stop()
	// Print initial line
	fmt.Fprintf(out, "%s: %s...\n", path, cases.Title(language.English).String(cmd.Action()))
	for {
		select {
		case <-ticker.C:
			elapsed += interval
			fmt.Fprintf(out, "%s: Still %s... [%s elapsed]\n", path, cmd.Action(), elapsed)
		case <-done:
			return
		}
	}
}
