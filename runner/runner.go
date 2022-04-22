package runner

import (
	"fmt"
	"os/exec"

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
)

func Run(config *parser.TerraConfig, opts ...func(r *TerraRun)) error {
	// Initialise TerraRun with defaults
	run := TerraRun{
		jobs: 1,
	}
	for _, opt := range opts {
		opt(&run)
	}

	// Start terraform runs in different root modules based on number of concurrent
	// jobs that are allowed
	swg := sizedwaitgroup.New(run.jobs)
	for _, tf := range config.RootModules() {
		swg.Add()
		tf := tf

		go func() {
			defer swg.Done()
			runCmds(&run, tf)
		}()
	}
	swg.Wait()

	return nil
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
	var args []string
	args = append(args,
		string(terraPlan),
		"-lock=true",
		"-input=false",
		"-out=tfplan",
	)
	args = append(args, run.extraArgs...)
	return runCmd(tf, args)
}

func applyCmd(run *TerraRun, tf *parser.Terrafile) error {
	var args []string
	args = append(args,
		string(terraApply),
		"-lock=true",
		"-input=false",
		"tfplan",
	)
	args = append(args, run.extraArgs...)
	return runCmd(tf, args)
}

func runCmd(tf *parser.Terrafile, args []string) error {
	args = append(tfArgs(tf), args...)
	execCmd := exec.Command(terraExe, args...)
	fmt.Printf("Executing:\n%s\n\n", execCmd.String())
	// stdout, err := execCmd.StdoutPipe()
	// if err != nil {
	// 	return fmt.Errorf("stdout pipe: %w", err)
	// }
	// var stderr bytes.Buffer
	// execCmd.Stderr = &stderr
	// // stderr, err := execCmd.StderrPipe()
	// // if err != nil {
	// // 	return fmt.Errorf("stderr pipe: %w", err)
	// // }
	// if startErr := execCmd.Start(); startErr != nil {
	// 	return fmt.Errorf("starting command: %w", err)
	// }

	// scanner := bufio.NewScanner(stdout)
	// scanner.Split(bufio.ScanWords)
	// for scanner.Scan() {
	// 	m := scanner.Text()
	// 	fmt.Println(m)
	// }

	// if waitErr := execCmd.Wait(); waitErr != nil {
	// 	return fmt.Errorf("waiting for command: %w", err)
	// }

	// if stderr.Len() > 0 {
	// 	fmt.Printf("%s\n", stderr.Bytes())
	// }

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
