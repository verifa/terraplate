package modulesui

import "github.com/charmbracelet/lipgloss"

type tableColumnsMap struct {
	status     column
	rootModule column
	summary    column
}

func (c tableColumnsMap) columns() []column {
	return []column{
		c.status, c.rootModule, c.summary,
	}
}

var (
	tableColumns = tableColumnsMap{
		status: column{
			header: "Status",
			width:  10,
			grow:   false,
		},
		rootModule: column{
			header: "Root Module",
			width:  60,
			grow:   true,
		},
		summary: column{
			header: "Summary",
			width:  60,
			grow:   true,
		},
	}
)

type column struct {
	header string
	width  int
	grow   bool
}

func (c column) render(style lipgloss.Style, text string) string {
	return style.Width(c.width).Render(text)
}
