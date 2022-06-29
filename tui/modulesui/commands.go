package modulesui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/verifa/terraplate/runner"
)

type SelectMsg struct {
	Run *runner.RunResult
}

func selectModuleCmd(run *runner.RunResult) tea.Cmd {
	return func() tea.Msg {
		return SelectMsg{Run: run}
	}
}

type BrowseMsg struct{}
