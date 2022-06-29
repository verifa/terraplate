package entryui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/verifa/terraplate/runner"
)

type BackMsg struct{}

func backToListMsg() tea.Cmd {
	return func() tea.Msg {
		return BackMsg{}
	}
}

type runFinished struct{}

type runInProgress struct{}

func RunInProgressCmd(module *runner.RootModule) tea.Cmd {

	var cmds = make([]tea.Cmd, 0)
	if module.IsRunning() {
		cmds = append(cmds, tickRunInProgress(module))
	}

	cmds = append(cmds, func() tea.Msg {
		module.Wait()
		return runFinished{}
	})

	return tea.Batch(cmds...)
}

func tickRunInProgress(module *runner.RootModule) tea.Cmd {
	return func() tea.Msg {
		time.Sleep(500 * time.Millisecond)
		return runInProgress{}
	}
}

func runModuleCmd(r *runner.Runner, module *runner.RootModule) tea.Cmd {
	r.Start([]*runner.RootModule{module})
	return RunInProgressCmd(module)
}
