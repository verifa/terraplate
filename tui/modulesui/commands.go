package modulesui

import (
	"errors"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/verifa/terraplate/runner"
)

type SelectMsg struct {
	Module *runner.RootModule
}

func (m Model) selectModuleCmd() tea.Cmd {
	return func() tea.Msg {
		return SelectMsg{Module: m.activeModule()}
	}
}

type updateModuleMsg struct {
	Module *runner.RootModule
}

func (m Model) runActiveModuleCmd() tea.Cmd {
	modules := []*runner.RootModule{
		m.activeModule(),
	}
	return m.runModulesWithOptsCmd(modules, m.runner.Opts)
}

func (m Model) runModulesWithOptsCmd(modules []*runner.RootModule, opts runner.TerraRunOpts) tea.Cmd {
	m.runner.StartWithOpts(modules, opts)

	var cmds = make([]tea.Cmd, len(modules))
	for _, mod := range modules {
		mod := mod
		cmds = append(cmds, func() tea.Msg {
			mod.Wait()
			return updateModuleMsg{Module: mod}
		})
	}

	return tea.Batch(cmds...)
}

func (m Model) inputCmd() (tea.Cmd, error) {
	opts, runAll, cmdErr := m.parseInputCmd(m.input.Value())
	if cmdErr != nil {
		return nil, cmdErr
	}
	var modules []*runner.RootModule
	if runAll {
		modules = m.visibleModules()
	} else {
		modules = append(modules, m.activeModule())
	}
	return m.runModulesWithOptsCmd(modules, opts), nil
}

func (m Model) parseInputCmd(cmd string) (runner.TerraRunOpts, bool, error) {
	var (
		opts   []func(*runner.TerraRunOpts)
		runAll bool
	)
	if len(cmd) == 0 {
		return runner.TerraRunOpts{}, runAll, errors.New("command is empty")
	}
	// Inherit from existing run, e.g. number of jobs and output
	opts = append(opts, runner.FromOpts(m.runner.Opts))
	for _, c := range cmd {
		switch c {
		case 'b':
			opts = append(opts, runner.RunBuild())
		case 'i':
			opts = append(opts, runner.RunBuild(), runner.RunInit())
		case 'u':
			opts = append(opts, runner.RunBuild(), runner.RunInitUpgrade())
		case 'p':
			// If we are planning, also run the terraform show command to get
			// the JSON output of the plan
			opts = append(opts, runner.RunBuild(), runner.RunPlan(), runner.RunShowPlan())
		case 'a':
			opts = append(opts, runner.RunApply())
		case 'A':
			runAll = true
		default:
			return runner.TerraRunOpts{}, runAll, fmt.Errorf("unknown command: %c", c)
		}
	}
	runOpts := runner.NewOpts(opts...)
	return runOpts, runAll, nil
}

type statusTimeoutMsg struct {
	// uuid uuid.UUID
	err error
}

func statusTimeoutCmd(err error) tea.Cmd {
	// uuid := uuid.New()
	timer := time.NewTimer(2 * time.Second)
	return func() tea.Msg {
		<-timer.C
		return statusTimeoutMsg{err: err}
	}
}
