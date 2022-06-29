package modulesui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

var (
	labelStyle = lipgloss.NewStyle().
			Padding(0, 1).MarginBottom(1)
	appliedLabel = labelStyle.Copy().
			Background(lipgloss.Color(termenv.ANSIGreen.String())).
			Render
	noChangesLabel = labelStyle.Copy().
			Background(lipgloss.Color(termenv.ANSIGreen.String())).
			Render
	driftLabel = labelStyle.Copy().
			Background(lipgloss.Color(termenv.ANSIYellow.String())).
			Render
	noPlanLabel = labelStyle.Copy().
			Background(lipgloss.Color(termenv.ANSICyan.String())).
			Render
	inProgress = labelStyle.Copy().
			Background(lipgloss.Color(termenv.ANSIMagenta.String())).
			Render
	errorLabel = labelStyle.Copy().
			Background(lipgloss.Color(termenv.ANSIRed.String())).
			Render

	boldStyle            = lipgloss.NewStyle().Bold(true)
	viewportContentStyle = lipgloss.NewStyle().Padding(0, 2)

	addSymbol = lipgloss.NewStyle().
			Foreground(lipgloss.Color(termenv.ANSIGreen.String())).
			Render("+")
	changeSymbol = lipgloss.NewStyle().
			Foreground(lipgloss.Color(termenv.ANSIYellow.String())).
			Render("~")
	destroySymbol = lipgloss.NewStyle().
			Foreground(lipgloss.Color(termenv.ANSIRed.String())).
			Render("-")
	replaceSymbol = destroySymbol + addSymbol

	dimmedColor = lipgloss.AdaptiveColor{Light: "#A49FA5", Dark: "#777777"}
)
