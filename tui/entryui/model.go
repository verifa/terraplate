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
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var _ tea.Model = (*Model)(nil)

func New(runner *runner.Runner, module *runner.RootModule, windowSize tea.WindowSizeMsg) tea.Model {

	m := Model{
		runner:     runner,
		activeTask: 0,
		help:       help.New(),
		module:     module,
		windowSize: windowSize,
	}

	m.viewport = viewport.New(0, 0)
	m.syncViewportSize()
	m.viewport.MouseWheelEnabled = true
	m.viewport.SetContent(m.viewportContent())

	return m
}

type Model struct {
	runner     *runner.Runner
	activeTask int
	viewport   viewport.Model
	help       help.Model
	module     *runner.RootModule
	windowSize tea.WindowSizeMsg
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.windowSize = msg
		m.syncViewportSize()
	case runInProgress:
		if m.module.IsRunning() {
			cmd = tickRunInProgress(m.module)
			m.viewport.SetContent(m.viewportContent())
			m.viewport.GotoBottom()
			m.syncViewportSize()
		}
	case runFinished:
		m.viewport.SetContent(m.viewportContent())
		m.viewport.GotoTop()
		m.syncViewportSize()
	case tea.KeyMsg:
		switch {
		case msg.String() == "q":
			cmd = backToListMsg()
		case key.Matches(msg, keys.back):
			cmd = backToListMsg()
		case key.Matches(msg, keys.nextSection):
			numTasks := m.numTasks()
			if numTasks > 0 {
				// We will use the index of this, so it has to be the length minus 1
				numTasks--
			}
			m.activeTask = min(m.activeTask+1, numTasks)
			m.viewport.SetContent(m.viewportContent())
			m.viewport.GotoTop()
		case key.Matches(msg, keys.prevSection):
			m.activeTask = max(m.activeTask-1, 0)
			m.viewport.SetContent(m.viewportContent())
			m.viewport.GotoTop()
		case key.Matches(msg, keys.run):
			cmd = runModuleCmd(m.runner, m.module)
			m.viewport.SetContent(m.viewportContent())
			m.viewport.GotoTop()
		default:
			m.viewport, cmd = m.viewport.Update(msg)
		}

	case tea.MouseMsg:
		m.viewport, cmd = m.viewport.Update(msg)
	}
	return m, cmd
}

func (m Model) View() string {
	return lipgloss.JoinVertical(
		lipgloss.Left,
		m.renderHeader(),
		m.viewport.View(),
		m.renderFooter(),
	)
}

func (m Model) viewportContent() string {
	style := lipgloss.NewStyle().Width(m.viewport.Width)
	run := m.module.Run
	switch {
	case run == nil:
		return style.Render("Not run.")
	case run.IsRunning():
		return style.Render(run.Log(true))
	case m.numTasks() == 0:
		return style.Render("No tasks to show")
	}
	return style.Render(run.Tasks[m.activeTask].Output.String())
}

func (m Model) renderHeader() string {
	var (
		run   = m.module.Run
		doc   strings.Builder
		tabs  []string
		caser = cases.Title(language.English)
	)
	// Don't render the tabs unless the run is finished
	if !m.module.IsRunning() {
		tabs = make([]string, m.numTasks())
		for i := 0; i < m.numTasks(); i++ {
			task := run.Tasks[i]
			if i == m.activeTask {
				tabs[i] = activeTab.Render(caser.String(string(task.TerraCmd)))
				continue
			}
			tabs[i] = tab.Render(caser.String(string(task.TerraCmd)))

		}
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
		m.renderStatus(),
		m.help.View(keys),
	)
}

func (m Model) renderStatus() string {
	return "Module: " + m.module.Terrafile.Dir
}

func (m *Model) syncViewportSize() {
	headerHeight := lipgloss.Height(m.renderHeader())
	footerHeight := lipgloss.Height(m.renderFooter())

	m.viewport.Width = m.windowSize.Width
	m.viewport.Height = max(0, m.windowSize.Height-headerHeight-footerHeight)
}

func (m Model) numTasks() int {
	if m.module.Run == nil {
		return 0
	}
	return len(m.module.Run.Tasks)
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
