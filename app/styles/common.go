package styles

import "github.com/charmbracelet/lipgloss"

var (
	Spacer = lipgloss.NewStyle().Width

	StatusBarLogCount = lipgloss.NewStyle().
				Padding(0, 1).
				Foreground(DefaultColor.Light).
				Render

	StatusBarBeamInfo = lipgloss.NewStyle().
				Padding(0, 1)

	ErrorInfo = lipgloss.NewStyle().
			Padding(0, 1).
			Background(DefaultColor.Error).
			Render
)
