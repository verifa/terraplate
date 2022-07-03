package modulesui

import (
	"github.com/charmbracelet/bubbles/key"
)

var delegateKeys = delegateKeyMap{
	choose: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "choose"),
	),
}

type delegateKeyMap struct {
	choose key.Binding
}

func (d delegateKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		d.choose,
	}
}

func (d delegateKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{
			d.choose,
		},
	}
}
