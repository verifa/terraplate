package runner

import (
	"io"
	"os"
)

func Jobs(jobs int) func(r *TerraRunOpts) {
	return func(r *TerraRunOpts) {
		r.jobs = jobs
	}
}

func RunBuild() func(r *TerraRunOpts) {
	return func(r *TerraRunOpts) {
		r.build = true
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

func RunInitUpgrade() func(r *TerraRunOpts) {
	return func(r *TerraRunOpts) {
		r.init = true
		r.initUpgrade = true
	}
}

func RunPlan() func(r *TerraRunOpts) {
	return func(r *TerraRunOpts) {
		r.plan = true
	}
}

func RunShow() func(r *TerraRunOpts) {
	return func(r *TerraRunOpts) {
		r.show = true
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

func Output(out io.Writer) func(r *TerraRunOpts) {
	return func(r *TerraRunOpts) {
		r.out = out
	}
}

func ExtraArgs(extraArgs []string) func(r *TerraRunOpts) {
	return func(r *TerraRunOpts) {
		r.extraArgs = extraArgs
	}
}

func FromOpts(opts TerraRunOpts) func(r *TerraRunOpts) {
	return func(r *TerraRunOpts) {
		r.jobs = opts.jobs
		r.out = opts.out
	}
}

func NewOpts(opts ...func(r *TerraRunOpts)) TerraRunOpts {
	// Initialise TerraRunOpts with defaults
	runOpts := TerraRunOpts{
		jobs: DefaultJobs,
	}
	for _, opt := range opts {
		opt(&runOpts)
	}

	// Set default output
	if runOpts.out == nil {
		runOpts.out = os.Stdout
	}
	return runOpts
}

// TerraRunOpts handles running Terraform over the root modules
type TerraRunOpts struct {
	out io.Writer

	build       bool
	validate    bool
	init        bool
	initUpgrade bool
	plan        bool
	show        bool
	showPlan    bool
	apply       bool

	// Max number of concurrent jobs allowed
	jobs int
	// Terraform command flags
	extraArgs []string
}
