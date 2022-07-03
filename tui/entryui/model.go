package entryui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/verifa/terraplate/runner"
)

var _ tea.Model = (*Model)(nil)

func New(run *runner.RunResult, windowSize tea.WindowSizeMsg) tea.Model {

	m := Model{
		activeTask: 0,
		totalTasks: len(run.Tasks),
		help:       help.New(),
		run:        run,
		windowSize: windowSize,
	}

	w, h := m.viewportSize()
	m.viewport = viewport.New(w, h)
	m.viewport.MouseWheelEnabled = true
	m.viewport.SetContent(m.activeTaskContent())

	return m
}

type Model struct {
	activeTask int
	totalTasks int
	viewport   viewport.Model
	help       help.Model
	run        *runner.RunResult
	windowSize tea.WindowSizeMsg
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.windowSize = msg
		w, h := m.viewportSize()
		m.viewport.Width = w
		m.viewport.Height = h
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.back):
			cmd = backToListMsg()
		case key.Matches(msg, keys.nextSection):
			m.activeTask = min(m.activeTask+1, m.totalTasks)
			m.viewport.SetContent(m.activeTaskContent())
		case key.Matches(msg, keys.prevSection):
			m.activeTask = max(m.activeTask-1, 0)
			m.viewport.SetContent(m.activeTaskContent())
			m.viewport.GotoTop()
		default:
			m.viewport, cmd = m.viewport.Update(msg)
		}

	case tea.MouseMsg:
		m.viewport, cmd = m.viewport.Update(msg)
	}
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	return lipgloss.JoinVertical(
		lipgloss.Left,
		// m.help.View(keys),
		m.renderHeader(),
		m.viewport.View(),
		m.renderFooter(),
	)
}

func (m Model) activeTaskContent() string {
	style := lipgloss.NewStyle().Width(m.viewport.Width)
	if len(m.run.Tasks) == 0 {
		return style.Render("No tasks to show")
	}
	if m.activeTask > len(m.run.Tasks) {
		return style.Render("Error: selected task out of range")
	}
	return style.Render(string(m.run.Tasks[m.activeTask].Output))
}

func (m Model) renderHeader() string {
	var (
		doc  strings.Builder
		tabs []string
	)
	tabs = make([]string, len(m.run.Tasks))
	for i, task := range m.run.Tasks {
		if i == m.activeTask {
			tabs[i] = activeTab.Render(strings.ToTitle(string(task.TerraCmd)))
			continue
		}
		tabs[i] = tab.Render(strings.ToTitle(string(task.TerraCmd)))
	}
	row := lipgloss.JoinHorizontal(lipgloss.Top, tabs...)
	gap := tabGap.Render(strings.Repeat(" ", max(0, m.windowSize.Width-lipgloss.Width(row))))
	doc.WriteString(lipgloss.JoinHorizontal(lipgloss.Bottom, row, gap))
	return doc.String() + "\n"
}

func (m Model) renderFooter() string {
	info := infoStyle.Render(fmt.Sprintf("%3.f%%", m.viewport.ScrollPercent()*100))
	line := strings.Repeat("─", max(0, m.viewport.Width-lipgloss.Width(info)))
	status := lipgloss.JoinHorizontal(
		lipgloss.Center,
		line,
		info,
	)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		status,
		m.help.View(keys),
	)
}

func (m Model) viewportSize() (int, int) {
	headerHeight := lipgloss.Height(m.renderHeader())
	footerHeight := lipgloss.Height(m.renderFooter())

	return m.windowSize.Width, max(0, m.windowSize.Height-headerHeight-footerHeight)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a > b {
		return b
	}
	return a
}
