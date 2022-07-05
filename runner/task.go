package runner

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

type TaskResult struct {
	ExecCmd  *exec.Cmd
	TerraCmd terraCmd

	Output  []byte
	Error   error
	Skipped bool
}

func (t *TaskResult) HasError() bool {
	return t.Error != nil
}

// HasRelevance is an attempt at better UX.
// We don't simply want to output everything. Things like successful inits and
// terraform show output are not interesting for the user, so skip them by
// default and therefore keep the output less
func (t *TaskResult) HasRelevance() bool {
	// Errors are always relevant
	if t.HasError() {
		return true
	}
	// Skipped tasks are not relevant
	if t.Skipped {
		return false
	}
	switch t.TerraCmd {
	case terraPlan:
		// Plan outputs are interesting
		return true
	case terraApply:
		// Apply outputs are interesting
		return true
	default:
		// Skip other command outputs
		return false
	}
}

func (t *TaskResult) Log() string {
	var summary strings.Builder

	summary.WriteString(fmt.Sprintf("%s output: %s\n\n", strings.Title(string(t.TerraCmd)), t.ExecCmd.String()))
	if t.HasError() {
		summary.WriteString(fmt.Sprintf("Error: %s\n\n", t.Error.Error()))
	}
	scanner := bufio.NewScanner(bytes.NewBuffer(t.Output))
	for scanner.Scan() {
		summary.WriteString(fmt.Sprintf("    %s\n", scanner.Text()))
	}
	summary.WriteString("\n\n")

	return summary.String()
}
