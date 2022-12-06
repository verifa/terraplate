package runner

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type TaskResult struct {
	ExecCmd  *exec.Cmd
	TerraCmd terraCmd

	Output  bytes.Buffer
	Error   error
	Skipped bool
}

func (t *TaskResult) HasError() bool {
	return t.Error != nil
}

func (t *TaskResult) Log() string {
	var (
		summary strings.Builder
		tmp     bytes.Buffer
		caser   = cases.Title(language.English)
	)

	// Make a copy of the output bytes as the scanner below will Read the io
	// and therefore "empty" it, and we don't want to empty the output bytes
	if _, err := tmp.Write(t.Output.Bytes()); err != nil {
		return "Error: writing task output to temporary buffer"
	}

	switch t.TerraCmd {
	case terraBuild:
		summary.WriteString("Build output:\n\n")
	default:
		summary.WriteString(fmt.Sprintf("%s output: %s\n\n", caser.String(string(t.TerraCmd)), t.ExecCmd.String()))
	}
	if t.HasError() {
		summary.WriteString(fmt.Sprintf("Error: %s\n\n", t.Error.Error()))
	}
	scanner := bufio.NewScanner(&tmp)
	for scanner.Scan() {
		summary.WriteString(fmt.Sprintf("    %s\n", scanner.Text()))
	}
	summary.WriteString("\n\n")

	return summary.String()
}
