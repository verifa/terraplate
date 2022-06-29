package modulesui

import (
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/verifa/terraplate/runner"
)

var _ tea.Model = (*Model)(nil)

func New(result *runner.Result) tea.Model {
	var m Model
	m.state = listView
	m.result = result

	// Setup list
	l := list.New(m.filterRows(), list.NewDefaultDelegate(), 0, 0)
	l.Title = "Root Modules"
	l.SetShowTitle(true)
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.AdditionalShortHelpKeys = keys.ShortHelp
	l.AdditionalFullHelpKeys = keys.FullHelpKeys
	m.list = l

	// Setup viewport
	m.viewport = viewport.New(0, 0)
	m.viewport.MouseWheelEnabled = true
	m.viewport.SetContent(m.viewportContent())

	m.showSummary = true
	m.showHelp = true
	m.listBorder = lipgloss.NewStyle().Border(lipgloss.NormalBorder())
	m.summaryBorder = lipgloss.NewStyle().Border(lipgloss.NormalBorder())
	m.setActiveViewStyle()

	return m
}

type Model struct {
	state    state
	result   *runner.Result
	list     list.Model
	viewport viewport.Model

	filterDrift bool
	filterError bool
	showSummary bool
	showHelp    bool

	listBorder    lipgloss.Style
	summaryBorder lipgloss.Style

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
		m.syncViewSizes()
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.filterDrift):
			m.filterDrift = !m.filterDrift
			m.list.SetItems(m.filterRows())
			m.viewport.SetContent(m.viewportContent())
			return m, nil
		case key.Matches(msg, keys.filterError):
			m.filterError = !m.filterError
			m.list.SetItems(m.filterRows())
			m.viewport.SetContent(m.viewportContent())
			return m, nil
		case key.Matches(msg, keys.toggleSummary):
			m.showSummary = !m.showSummary
			m.syncViewSizes()
			return m, nil
		case key.Matches(msg, keys.tab):
			if m.state == listView {
				m.state = summaryView
			} else {
				m.state = listView
			}
			m.setActiveViewStyle()
			return m, nil
		case key.Matches(msg, keys.toggleHelpMenu):
			m.showHelp = !m.showHelp
			m.syncViewSizes()
			return m, nil
		case key.Matches(msg, keys.enter):
			cmd = selectModuleCmd(m.activeModule())
		default:
			switch m.state {
			case listView:
				// Call the list delegate, and check if the selected item has changed.
				// If the selected item has changed, update the viewport to show the
				// correct summary
				prevItem := m.list.SelectedItem()
				m.list, cmd = m.list.Update(msg)
				newItem := m.list.SelectedItem()
				if prevItem != newItem {
					m.viewport.SetContent(m.viewportContent())
				}

				// If the short/long help was toggled, we need to re-sync the views
				if key.Matches(msg, m.list.KeyMap.CloseFullHelp, m.list.KeyMap.ShowFullHelp) {
					m.syncViewSizes()
				}
			case summaryView:
				m.viewport, cmd = m.viewport.Update(msg)
			}
		}
		cmds = append(cmds, cmd)
	}
	return m, tea.Batch(cmds...)
}
func (m Model) View() string {
	views := []string{
		m.listBorder.Render(m.list.View()),
	}
	if m.showSummary {
		views = append(views, m.summaryBorder.Render(m.viewport.View()))
	}
	m.list.Help.View(m.list)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		// m.renderHeader(),
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			views...,
		),
		m.renderFooter(),
	)
}

func (m Model) filterRows() []list.Item {
	var rows = make([]list.Item, 0)
	for _, run := range m.result.Runs {
		drift := run.Drift()
		if m.filterDrift && !drift.HasDrift() {
			continue
		}
		if m.filterError && !run.HasError() {
			continue
		}
		rows = append(rows, listItem{
			run: run,
		})
	}
	return rows
}

func (m Model) activeModule() *runner.RunResult {
	items := m.list.Items()
	activeItem := items[m.list.Index()]
	return activeItem.(listItem).run
}

func (m *Model) setActiveViewStyle() {
	if m.state == listView {
		m.summaryBorder.BorderForeground(dimmedColor)
		m.listBorder.UnsetBorderForeground()
	} else {
		m.listBorder.BorderForeground(dimmedColor)
		m.summaryBorder.UnsetBorderForeground()
	}

}

func (m *Model) syncViewSizes() {
	var (
		footerHeight = lipgloss.Height(m.renderFooter())
		lvW          int
		lvH          = m.windowSize.Height - footerHeight
		svW          int
		svH          = m.windowSize.Height - footerHeight
	)
	if m.showSummary {
		svW = m.windowSize.Width / 2
		lvW = m.windowSize.Width - svW
	} else {
		lvW = m.windowSize.Width
		svW = 0
	}

	m.listBorder.Width(lvW - m.listBorder.GetHorizontalBorderSize())
	m.listBorder.Height(lvH - m.listBorder.GetVerticalBorderSize())
	m.summaryBorder.Width(svW - m.summaryBorder.GetHorizontalBorderSize())
	m.summaryBorder.Height(svH - m.summaryBorder.GetVerticalBorderSize())

	listWidth := lvW - m.listBorder.GetHorizontalFrameSize()
	listHeight := lvH - m.listBorder.GetVerticalFrameSize()
	m.list.SetSize(listWidth, listHeight)

	summaryWidth := svW - m.summaryBorder.GetHorizontalFrameSize()
	summaryHeight := svH - m.summaryBorder.GetVerticalFrameSize()
	m.viewport.Width = summaryWidth
	m.viewport.Height = summaryHeight
}

func (m Model) viewportContent() string {
	return viewportContentStyle.Render(
		m.renderSummaryContent(m.list.SelectedItem()),
	)
}

func (m Model) renderFooter() string {
	if m.showHelp {
		return m.list.Help.View(m.list)
	}
	return ""
}

func (m Model) renderSummaryContent(item list.Item) string {
	if item == nil {
		return "No module selected."
	}

	var (
		s                strings.Builder
		hasDrift         bool
		noOpResources    []*tfjson.ResourceChange
		addResources     []*tfjson.ResourceChange
		changeResources  []*tfjson.ResourceChange
		replaceResources []*tfjson.ResourceChange
		destroyResources []*tfjson.ResourceChange
	)

	run := item.(listItem).run
	// Initialize tabwriter for making pretty tab-indented text
	tw := tabwriter.NewWriter(&s, 0, 0, 2, ' ', tabwriter.DiscardEmptyColumns)

	if !run.HasPlan() {
		fmt.Fprintln(&s, noPlanLabel("No plan"))
		fmt.Fprintf(&s, "No plan available.")
		return s.String()
	}
	if run.HasError() {
		fmt.Fprintln(&s, errorLabel("Error"))
		fmt.Fprintf(&s, "Error running Terraform.")
		return s.String()

	}
	for _, r := range run.Plan.ResourceChanges {
		actions := r.Change.Actions
		switch {
		case actions.NoOp():
			noOpResources = append(noOpResources, r)
		case actions.Create():
			hasDrift = true
			addResources = append(addResources, r)
		case actions.Update():
			hasDrift = true
			changeResources = append(changeResources, r)
		case actions.Replace():
			hasDrift = true
			replaceResources = append(replaceResources, r)
		case actions.Delete():
			hasDrift = true
			destroyResources = append(destroyResources, r)
		}
	}

	// Header
	if hasDrift {
		fmt.Fprintln(&s, driftLabel("Drift detected"))
	} else {
		fmt.Fprintln(&s, noChangesLabel("No changes"))
	}

	// Overview
	fmt.Fprintf(&s, "%s\n\n", boldStyle.Render("Overview"))
	fmt.Fprintf(tw, "Terraform version\t%s\n", run.Plan.TerraformVersion)
	fmt.Fprintf(tw, "Total resources\t%d\n", len(run.Plan.ResourceChanges))
	tw.Flush()

	// Plan summary
	fmt.Fprintf(&s, "\n%s\n", boldStyle.Render("Plan summary"))
	if !hasDrift {
		fmt.Fprintln(&s, "No changes.")
	}
	if len(addResources) != 0 {
		fmt.Fprintf(&s, "\nAdd:\n")
		for _, r := range addResources {
			fmt.Fprintf(&s, "%s %s\n", addSymbol, r.Address)
		}
	}
	if len(changeResources) != 0 {
		fmt.Fprintf(&s, "\nChange:\n")
		for _, r := range changeResources {
			fmt.Fprintf(&s, "%s %s\n", changeSymbol, r.Address)
		}
	}
	if len(replaceResources) != 0 {
		fmt.Fprintf(&s, "\nReplace:\n")
		for _, r := range replaceResources {
			fmt.Fprintf(&s, "%s %s\n", replaceSymbol, r.Address)
		}
	}
	if len(destroyResources) != 0 {
		fmt.Fprintf(&s, "\nDestroy:\n")
		for _, r := range destroyResources {
			fmt.Fprintf(&s, "%s %s\n", destroySymbol, r.Address)
		}
	}

	// Modules
	fmt.Fprintf(&s, "\n%s\n\n", boldStyle.Render("Modules"))
	if run.Plan.Config.RootModule.ModuleCalls == nil {
		fmt.Fprintln(&s, "No modules calls.")
	}
	for name, mc := range run.Plan.Config.RootModule.ModuleCalls {
		fmt.Fprintf(tw, "%s\t%s\n", name, mc.Source)
		tw.Flush()
	}
	// Providers
	fmt.Fprintf(&s, "\n%s\n\n", boldStyle.Render("Providers"))
	for name, config := range run.Plan.Config.ProviderConfigs {
		var pName string
		if config.Alias != "" {
			pName = fmt.Sprintf("%s (%s)", name, config.Alias)
		} else {
			pName = name
		}
		fmt.Fprintf(tw, "%s\t%s\n", pName, config.VersionConstraint)
		tw.Flush()
	}

	return s.String()
}
