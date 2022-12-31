package styles

import "github.com/charmbracelet/lipgloss"

const (
	ColorLogo              = lipgloss.Color("93")
	ColorBorder            = lipgloss.Color("0")
	ColorBorderActive      = ColorLogo // tight to the logo color
	ColorStatusBarLogCount = lipgloss.Color("93")
	ColorStatusBarBeamInfo = lipgloss.Color("22")
)

var (
	StatusBarLogCount = lipgloss.NewStyle().
				Padding(0, 1).
				Background(ColorStatusBarLogCount).
				Render

	StatusBarBeamInfo = lipgloss.NewStyle().
				Padding(0, 1).
				Background(ColorStatusBarBeamInfo).
				Render
)
