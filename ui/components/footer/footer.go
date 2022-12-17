package footer

import (
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	width, height int

	mode string

	streamCount int
	logCount    int
}

func New(width, height int) Model {
	return Model{
		width:       width,
		height:      height,
		mode:        "LOGGING",
		streamCount: 3,
		logCount:    124,
	}
}

var (
	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#343433", Dark: "#C1C6B2"}).
			Background(lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#353533"})

	spacer = lipgloss.NewStyle().Width

	modeStyle = lipgloss.NewStyle().
			Inherit(statusBarStyle).
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#FF5F87")).
			Padding(0, 1)

	statusNugget = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Padding(0, 1)

	nuggetValue = statusNugget.Copy().Bold(true)
)

func (m Model) View() string {
	return lipgloss.JoinHorizontal(lipgloss.Left,
		modeStyle.Render(m.mode),
	)
}
