package modulesui

import "github.com/charmbracelet/bubbles/key"

var keys = keymap{
	enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select"),
	),
	colon: key.NewBinding(
		key.WithKeys(":"),
		key.WithHelp(":", "command mode"),
	),
	esc: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back to list"),
	),
	run: key.NewBinding(
		key.WithKeys(" "),
		key.WithHelp("␣", "run selected"),
	),
	// runAll: key.NewBinding(
	// 	key.WithKeys("ctrl+r"),
	// 	key.WithHelp("ctrl+r", "run all visible"),
	// ),
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
	enter key.Binding
	esc   key.Binding
	run   key.Binding
	// runAll           key.Binding
	tab              key.Binding
	colon            key.Binding
	filterDrift      key.Binding
	filterError      key.Binding
	toggleSummary    key.Binding
	togglePagination key.Binding
	toggleHelpMenu   key.Binding
}

func (k keymap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.enter, k.colon, k.run, k.tab, k.filterDrift,
	}
}

func (k keymap) FullHelpKeys() []key.Binding {
	return []key.Binding{
		k.enter, k.colon, k.run, k.tab, k.filterDrift, k.filterError, k.toggleSummary,
		k.togglePagination, k.toggleHelpMenu,
	}
}

var inputKeys = inputKeyMap{
	build: key.NewBinding(
		key.WithKeys("b"),
		key.WithHelp("b", "build"),
	),
	init: key.NewBinding(
		key.WithKeys("i"),
		key.WithHelp("i", "init"),
	),
	upgrade: key.NewBinding(
		key.WithKeys("u"),
		key.WithHelp("u", "upgrade"),
	),
	plan: key.NewBinding(
		key.WithKeys("p"),
		key.WithHelp("p", "plan"),
	),
	apply: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "apply"),
	),
	all: key.NewBinding(
		key.WithKeys("A"),
		key.WithHelp("A", "all modules"),
	),
}

type inputKeyMap struct {
	build   key.Binding
	init    key.Binding
	upgrade key.Binding
	plan    key.Binding
	apply   key.Binding
	all     key.Binding
}

func (k inputKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.build, k.init, k.upgrade, k.plan, k.apply, k.all,
	}
}
func (k inputKeyMap) FullHelp() [][]key.Binding {
	return nil
}
