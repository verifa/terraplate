package entryui

import (
	tea "github.com/charmbracelet/bubbletea"
)

type BackMsg struct{}

func backToListMsg() tea.Cmd {
	return func() tea.Msg {
		return BackMsg{}
	}
}
