package runner

import (
	"fmt"
	"os/exec"

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
	var run TerraRun
	for _, opt := range opts {
		opt(&run)
	}

	for _, tf := range config.RootModules() {
		fmt.Println("")
		fmt.Println("##################################")
		fmt.Println("Calling Runner in", tf.Dir)
		fmt.Println("##################################")
		fmt.Println("")
		if run.init {
			if err := initCmd(&run, tf); err != nil {
				return fmt.Errorf("terrafile %s: %w", tf.Path, err)
			}
		}
		if run.validate {
			if err := validateCmd(&run, tf); err != nil {
				return fmt.Errorf("terrafile %s: %w", tf.Path, err)
			}
		}
		if run.plan {
			if err := planCmd(&run, tf); err != nil {
				return fmt.Errorf("terrafile %s: %w", tf.Path, err)
			}
		}
		if run.apply {
			if err := applyCmd(&run, tf); err != nil {
				return fmt.Errorf("terrafile %s: %w", tf.Path, err)
			}
		}
	}

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

func InitUpgrade(upgrade bool) func(r *TerraRun) {
	return func(r *TerraRun) {
		r.upgrade = upgrade
	}
}

type TerraRun struct {
	validate bool
	init     bool
	plan     bool
	apply    bool

	// Terraform command flags
	upgrade bool
}

func initCmd(run *TerraRun, tf *parser.Terrafile) error {
	var args []string
	args = append(args, string(terraInit))
	if run.upgrade {
		args = append(args, "--upgrade")
	}
	return runCmd(tf, args)
}

func validateCmd(run *TerraRun, tf *parser.Terrafile) error {
	var args []string
	args = append(args, string(terraValidate))
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
