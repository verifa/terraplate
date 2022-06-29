package modulesui

import "github.com/charmbracelet/bubbles/key"

var keys = keymap{
	enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select"),
	),
	tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "select section"),
	),
	filterDrift: key.NewBinding(
		key.WithKeys("D"),
		key.WithHelp("D", "filter drift"),
	),
	filterError: key.NewBinding(
		key.WithKeys("E"),
		key.WithHelp("E", "filter error"),
	),
	toggleSummary: key.NewBinding(
		key.WithKeys("S"),
		key.WithHelp("S", "toggle summary"),
	),
	togglePagination: key.NewBinding(
		key.WithKeys("P"),
		key.WithHelp("P", "toggle pagination"),
	),
	toggleHelpMenu: key.NewBinding(
		key.WithKeys("H"),
		key.WithHelp("H", "toggle help"),
	),
}

type keymap struct {
	enter            key.Binding
	tab              key.Binding
	filterDrift      key.Binding
	filterError      key.Binding
	toggleSummary    key.Binding
	togglePagination key.Binding
	toggleHelpMenu   key.Binding
}

func (k keymap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.enter, k.tab, k.filterDrift,
	}
}

func (k keymap) FullHelpKeys() []key.Binding {
	return []key.Binding{
		k.enter, k.tab, k.filterDrift, k.filterError, k.toggleSummary,
		k.togglePagination, k.toggleHelpMenu,
	}
}
