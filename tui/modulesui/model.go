package modulesui

import (
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/verifa/terraplate/runner"
)

var _ tea.Model = (*Model)(nil)

func New(runner *runner.Runner) tea.Model {
	var m Model
	m.state = listView
	m.runner = runner

	// Setup list
	l := list.New(m.filterItems(), list.NewDefaultDelegate(), 0, 0)
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

	// Setup textinput for commands
	input := textinput.New()
	input.Prompt = ":"
	m.input = input

	return m
}

type Model struct {
	state     state
	runner    *runner.Runner
	list      list.Model
	viewport  viewport.Model
	input     textinput.Model
	statusErr error

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
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.windowSize = msg
		m.syncViewSizes()
	case updateModuleMsg:
		// If the selected module is updated, update the viewport
		if msg.Module == m.activeModule() {
			m.viewport.SetContent(m.viewportContent())
			m.viewport.GotoTop()
		}
	case statusTimeoutMsg:
		if msg.err == m.statusErr {
			m.statusErr = nil
		}

	case list.FilterMatchesMsg:
		cmd = m.handleListUpdate(msg)
	case tea.KeyMsg:
		if m.list.FilterState() == list.Filtering {
			// If we are filtering the list, send all keys to the list
			cmd = m.handleListUpdate(msg)
		} else if m.input.Focused() {
			switch {
			case key.Matches(msg, keys.enter):
				var cmdErr error
				cmd, cmdErr = m.inputCmd()
				if cmdErr != nil {
					m.statusErr = cmdErr
					cmd = statusTimeoutCmd(cmdErr)
				}
				m.input.Blur()
				m.input.Reset()
				m.syncViewSizes()
				m.setActiveViewStyle()
			case key.Matches(msg, keys.esc):
				m.input.Blur()
				m.input.Reset()
				m.syncViewSizes()
				m.setActiveViewStyle()
			default:
				m.input, cmd = m.input.Update(msg)
			}
		} else {
			switch {
			case key.Matches(msg, keys.enter):
				cmd = m.selectModuleCmd()
			case key.Matches(msg, keys.colon):
				cmd = m.input.Focus()
				m.syncViewSizes()
				m.setActiveViewStyle()
			case key.Matches(msg, keys.esc):
				switch m.state {
				case listView:
					cmd = m.handleListUpdate(msg)
				case summaryView:
					m.state = listView
				}
				m.setActiveViewStyle()
			case key.Matches(msg, keys.run):
				cmd = m.runActiveModuleCmd()
				m.viewport.SetContent(m.viewportContent())
				m.viewport.GotoTop()
			case key.Matches(msg, keys.tab):
				if m.state == listView {
					m.state = summaryView
				} else {
					m.state = listView
				}
				m.setActiveViewStyle()
				return m, nil
			case key.Matches(msg, keys.filterDrift):
				m.filterDrift = !m.filterDrift
				m.list.SetItems(m.filterItems())
				// Reset the list filter. Slightly strange behaviour if we don't
				// do this.
				m.list.ResetFilter()
				m.viewport.SetContent(m.viewportContent())
				m.viewport.GotoTop()
				return m, nil
			case key.Matches(msg, keys.filterError):
				m.filterError = !m.filterError
				m.list.SetItems(m.filterItems())
				// Reset the list filter. Slightly strange behaviour if we don't
				// do this.
				m.list.ResetFilter()
				m.viewport.SetContent(m.viewportContent())
				m.viewport.GotoTop()
				return m, nil
			case key.Matches(msg, keys.togglePagination):
				m.list.SetShowPagination(!m.list.ShowPagination())
			case key.Matches(msg, keys.toggleSummary):
				m.showSummary = !m.showSummary
				m.state = listView
				m.setActiveViewStyle()
				m.syncViewSizes()
				return m, nil
			case key.Matches(msg, keys.toggleHelpMenu):
				m.showHelp = !m.showHelp
				m.syncViewSizes()
				return m, nil

			default:
				if key.Matches(msg, m.list.KeyMap.CloseFullHelp, m.list.KeyMap.ShowFullHelp) {
					cmd = m.handleListUpdate(msg)
					m.syncViewSizes()
					break
				}
				switch m.state {
				case listView:
					cmd = m.handleListUpdate(msg)
				case summaryView:
					m.viewport, cmd = m.viewport.Update(msg)
				}
			}
		}
	}
	return m, cmd
}

func (m Model) View() string {

	views := []string{
		m.listBorder.Render(m.list.View()),
	}
	if m.showSummary {
		views = append(views, m.summaryBorder.Render(m.viewport.View()))
	}

	stacks := []string{
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			views...,
		),
	}
	if m.input.Focused() {
		stacks = append(stacks, m.input.View())
	} else {
		stacks = append(stacks, m.renderStatus())
	}
	stacks = append(stacks, m.renderFooter())

	return lipgloss.JoinVertical(
		lipgloss.Left,
		stacks...,
	)
}

func (m *Model) handleListUpdate(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	// Call the list delegate, and check if the selected item has changed.
	// If the selected item has changed, update the viewport to show the
	// correct summary
	prevItem := m.list.SelectedItem()
	prevState := m.list.FilterState()
	m.list, cmd = m.list.Update(msg)
	newItem := m.list.SelectedItem()
	newState := m.list.FilterState()
	if prevItem != newItem || prevState != newState {
		m.viewport.SetContent(m.viewportContent())
		m.viewport.GotoTop()
	}
	return cmd
}

func (m Model) filterItems() []list.Item {
	var items = make([]list.Item, 0)
	for _, mod := range m.runner.Modules {
		var (
			run   = mod.Run
			drift = run.Drift()
		)
		if m.filterDrift && !drift.HasDrift() {
			continue
		}
		if m.filterError && !run.HasError() {
			continue
		}
		items = append(items, listItem{
			module: mod,
		})
	}
	return items
}

func (m Model) activeModule() *runner.RootModule {
	item := m.list.SelectedItem()
	if item == nil {
		return nil
	}
	return item.(listItem).module
}

func (m Model) visibleModules() []*runner.RootModule {
	items := m.list.VisibleItems()
	modules := make([]*runner.RootModule, len(items))
	for index, item := range items {
		modules[index] = item.(listItem).module
	}
	return modules
}

func (m *Model) setActiveViewStyle() {
	if m.input.Focused() {
		m.summaryBorder.BorderForeground(dimmedColor)
		m.listBorder.BorderForeground(dimmedColor)
		return
	}
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
		footerHeight                        = lipgloss.Height(m.renderFooter())
		statusHeight                        int
		listViewWidth, listViewHeight       int
		summaryViewWidth, summaryViewHeight int
	)
	if m.input.Focused() {
		statusHeight = lipgloss.Height(m.input.View())
	} else {
		statusHeight = 1
	}
	listViewHeight = m.windowSize.Height - footerHeight - statusHeight
	summaryViewHeight = listViewHeight
	if m.showSummary {
		summaryViewWidth = m.windowSize.Width / 2
		listViewWidth = m.windowSize.Width - summaryViewWidth
	} else {
		listViewWidth = m.windowSize.Width
		summaryViewWidth = 0
	}

	m.listBorder.Width(listViewWidth - m.listBorder.GetHorizontalBorderSize())
	m.listBorder.Height(listViewHeight - m.listBorder.GetVerticalBorderSize())
	m.summaryBorder.Width(summaryViewWidth - m.summaryBorder.GetHorizontalBorderSize())
	m.summaryBorder.Height(summaryViewHeight - m.summaryBorder.GetVerticalBorderSize())

	listWidth := listViewWidth - m.listBorder.GetHorizontalFrameSize()
	listHeight := listViewHeight - m.listBorder.GetVerticalFrameSize()
	m.list.SetSize(listWidth, listHeight)

	summaryWidth := summaryViewWidth - m.summaryBorder.GetHorizontalFrameSize()
	summaryHeight := summaryViewHeight - m.summaryBorder.GetVerticalFrameSize()
	m.viewport.Width = summaryWidth
	m.viewport.Height = summaryHeight
	viewportContentStyle.Width(summaryWidth)
}

func (m Model) viewportContent() string {
	var content string
	if m.list.FilterState() == list.Filtering {
		content = "Filtering..."
	} else {
		content = m.renderSummaryContent(m.list.SelectedItem())
	}
	return viewportContentStyle.Render(content)
}

func (m Model) renderStatus() string {
	if m.statusErr != nil {
		return "Error: " + m.statusErr.Error()
	}
	return "Base: " + m.runner.WorkingDirectory()
}

func (m Model) renderFooter() string {
	if m.input.Focused() {
		return help.New().View(inputKeys)
	}
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

	module := item.(listItem).module
	run := module.Run
	// Initialize tabwriter for making pretty tab-indented text
	tw := tabwriter.NewWriter(&s, 0, 0, 2, ' ', tabwriter.DiscardEmptyColumns)

	if run.IsRunning() {
		fmt.Fprintln(&s, inProgress("Running..."))
		fmt.Fprintf(&s, "Run is in progress...")
		return s.String()
	}
	if run.IsApplied() {
		fmt.Fprintln(&s, appliedLabel("Applied"))
		fmt.Fprintf(&s, "Terraform applied.")
		return s.String()
	}
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
