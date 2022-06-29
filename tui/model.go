package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/verifa/terraplate/runner"
	"github.com/verifa/terraplate/tui/entryui"
	"github.com/verifa/terraplate/tui/modulesui"
)

var _ tea.Model = (*MainModel)(nil)

type MainModel struct {
	runner *runner.Runner

	modules tea.Model
	entry   tea.Model

	state      state
	windowSize tea.WindowSizeMsg
}

func New(runner *runner.Runner) MainModel {
	return MainModel{
		runner:  runner,
		state:   modulesView,
		modules: modulesui.New(runner),
	}
}

func (m MainModel) Init() tea.Cmd {
	return nil
}

func (m MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.windowSize = msg // pass this along to the entry view so it uses the full window size when it's initialized
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}
	case entryui.BackMsg:
		m.state = modulesView
	case modulesui.SelectMsg:
		m.state = entryView
		m.entry = entryui.New(m.runner, msg.Module, m.windowSize)
		if msg.Module.IsRunning() {
			cmds = append(cmds, entryui.RunInProgressCmd(msg.Module))
		}
	}

	switch m.state {
	case modulesView:
		m.modules, cmd = m.modules.Update(msg)
	case entryView:
		m.entry, cmd = m.entry.Update(msg)
	}
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m MainModel) View() string {
	switch m.state {
	case modulesView:
		return m.modules.View()
	case entryView:
		return m.entry.View()
	}
	return "Error: no view selected"
}
