package runner

import (
	"fmt"
	"os/exec"

	"github.com/hashicorp/go-multierror"
	"github.com/remeh/sizedwaitgroup"
	"github.com/verifa/terraplate/parser"
)

type terraCmd string

const (
	terraExe               = "terraform"
	terraValidate terraCmd = "validate"
	terraInit     terraCmd = "init"
	terraPlan     terraCmd = "plan"
	terraApply    terraCmd = "apply"

	DefaultJobs = 1
)

func Run(config *parser.TerraConfig, opts ...func(r *TerraRun)) error {
	// Initialise TerraRun with defaults
	run := TerraRun{
		jobs: DefaultJobs,
	}
	for _, opt := range opts {
		opt(&run)
	}

	var errors error
	// Start terraform runs in different root modules based on number of concurrent
	// jobs that are allowed
	swg := sizedwaitgroup.New(run.jobs)
	for _, tf := range config.RootModules() {
		swg.Add()
		tf := tf

		go func() {
			defer swg.Done()
			if err := runCmds(&run, tf); err != nil {
				errors = multierror.Append(errors, err)
			}
		}()
	}
	swg.Wait()

	return errors
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
	apply    bool

	// Max number of concurrent jobs allowed
	jobs int
	// Terraform command flags
	extraArgs []string
}

func runCmds(run *TerraRun, tf *parser.Terrafile) error {
	// Check if root module should be skipped or not
	if tf.ExecBlock.Skip {
		fmt.Println("")
		fmt.Println("Skipping runner for", tf.Dir)
		fmt.Println("")
		return nil
	}
	fmt.Println("")
	fmt.Println("##################################")
	fmt.Println("Calling Runner in", tf.Dir)
	fmt.Println("##################################")
	fmt.Println("")
	if run.init {
		if err := initCmd(run, tf); err != nil {
			return fmt.Errorf("terrafile %s: %w", tf.Path, err)
		}
	}
	if run.validate {
		if err := validateCmd(run, tf); err != nil {
			return fmt.Errorf("terrafile %s: %w", tf.Path, err)
		}
	}
	if run.plan {
		if err := planCmd(run, tf); err != nil {
			return fmt.Errorf("terrafile %s: %w", tf.Path, err)
		}
	}
	if run.apply {
		if err := applyCmd(run, tf); err != nil {
			return fmt.Errorf("terrafile %s: %w", tf.Path, err)
		}
	}
	return nil
}

func initCmd(run *TerraRun, tf *parser.Terrafile) error {
	var args []string
	args = append(args, string(terraInit))
	args = append(args, run.extraArgs...)
	return runCmd(tf, args)
}

func validateCmd(run *TerraRun, tf *parser.Terrafile) error {
	var args []string
	args = append(args, string(terraValidate))
	args = append(args, run.extraArgs...)
	return runCmd(tf, args)
}

func planCmd(run *TerraRun, tf *parser.Terrafile) error {
	plan := tf.ExecBlock.PlanBlock

	var args []string
	args = append(args,
		string(terraPlan),
		fmt.Sprintf("-lock=%v", plan.Lock),
		fmt.Sprintf("-input=%v", plan.Input),
	)
	if !plan.SkipOut {
		args = append(args,
			"-out="+plan.Out,
		)
	}
	args = append(args, tf.ExecBlock.ExtraArgs...)
	args = append(args, run.extraArgs...)
	return runCmd(tf, args)
}

func applyCmd(run *TerraRun, tf *parser.Terrafile) error {
	plan := tf.ExecBlock.PlanBlock

	var args []string
	args = append(args,
		string(terraApply),
		fmt.Sprintf("-lock=%v", plan.Lock),
		fmt.Sprintf("-input=%v", plan.Input),
	)
	args = append(args, tf.ExecBlock.ExtraArgs...)
	args = append(args, run.extraArgs...)

	if !plan.SkipOut {
		args = append(args, plan.Out)
	}

	return runCmd(tf, args)
}

func runCmd(tf *parser.Terrafile, args []string) error {
	args = append(tfArgs(tf), args...)
	execCmd := exec.Command(terraExe, args...)
	fmt.Printf("Executing:\n%s\n\n", execCmd.String())

	out, runErr := execCmd.CombinedOutput()
	if runErr != nil {
		return fmt.Errorf("executing command %s: %w\n%s", execCmd.String(), runErr, out)
	}
	fmt.Printf("%s\n", out)
	return nil
}

func tfArgs(tf *parser.Terrafile) []string {
	var args []string
	args = append(args, "-chdir="+tf.Dir)
	return args
}
