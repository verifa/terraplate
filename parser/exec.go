package parser

type ExecBlock struct {
	ExtraArgs []string       `hcl:"extra_args,optional"`
	PlanBlock *ExecPlanBlock `hcl:"plan,block"`
}

// ExecPlanBlock defines the arguments for running the Terraform plan
type ExecPlanBlock struct {
	Input bool `hcl:"input,optional"`
	Lock  bool `hcl:"lock,optional"`
	// Out specifies the name of the Terraform plan file to create (if any)
	Out string `hcl:"out,optional"`
	// SkipOut says whether to avoid creating a plan file or not. Useful in cases
	// such as running Terraform Cloud remotely, where plans cannot be created.
	SkipOut bool `hcl:"skip_out,optional"`
}
