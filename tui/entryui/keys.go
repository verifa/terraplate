package entryui

import "github.com/charmbracelet/bubbles/key"

var keys = keymap{
	prevSection: key.NewBinding(
		key.WithKeys("left", "h"),
		key.WithHelp("/h", "previous section"),
	),
	nextSection: key.NewBinding(
		key.WithKeys("right", "l"),
		key.WithHelp("/l", "next section"),
	),
	back: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back to list"),
	),
	run: key.NewBinding(
		key.WithKeys(" "),
		key.WithHelp("␣", "run selected"),
	),
}

type keymap struct {
	back        key.Binding
	nextSection key.Binding
	prevSection key.Binding
	run         key.Binding
}

func (k keymap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.back, k.nextSection, k.prevSection, k.run,
	}
}

func (k keymap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{
			k.back, k.nextSection, k.prevSection, k.run,
		},
	}
}
