package welcome

import (
	"math"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	welcomeLogo = lipgloss.NewStyle().
			MarginBottom(2).
			Render(
			strings.Join([]string{
				lipgloss.NewStyle().Foreground(lipgloss.Color("#FF4C94")).Render(
					"███████╗ ██████╗ ██████╗ ████████╗████████╗██╗   ██╗",
				),
				lipgloss.NewStyle().Foreground(lipgloss.Color("#EF46AC")).Render(
					"██╔════╝██╔════╝██╔═══██╗╚══██╔══╝╚══██╔══╝╚██╗ ██╔╝",
				),
				lipgloss.NewStyle().Foreground(lipgloss.Color("#D840C0")).Render(
					"███████╗██║     ██║   ██║   ██║      ██║    ╚████╔╝ ",
				),
				lipgloss.NewStyle().Foreground(lipgloss.Color("#BE38D5")).Render(
					"╚════██║██║     ██║   ██║   ██║      ██║     ╚██╔╝ ",
				),
				lipgloss.NewStyle().Foreground(lipgloss.Color("#BE38D5")).Render(
					"███████║╚██████╗╚██████╔╝   ██║      ██║      ██║",
				),
				lipgloss.NewStyle().Foreground(lipgloss.Color("#9F2DEB")).Render(
					"╚══════╝ ╚═════╝ ╚═════╝    ╚═╝      ╚═╝      ╚═╝",
				),
			}, "\n"),
		)

	welcomeUsage = lipgloss.NewStyle().
			MarginBottom(2).
			Render(
			strings.Join([]string{
				lipgloss.NewStyle().Bold(true).Underline(true).Render("beam logs:\n"),
				"\tfrom stderr: " + lipgloss.NewStyle().Bold(false).Render("go run -race my/awesome/app.go 2>&1 | beam -label=navigation_service"),
				"\tfrom stdout: " + lipgloss.NewStyle().Bold(false).Render("cat uss_enterprise_engine_logs.log | beam -label=engine_service"),
			}, "\n"),
		)

	howToText = lipgloss.NewStyle().
			Render(
			strings.Join([]string{
				lipgloss.NewStyle().Bold(true).Underline(true).Render("tips and notes:\n"),
				lipgloss.NewStyle().Bold(false).Render("\t- hit \":\" and type an index to format a specific line. Use k/j to format the previous or next log"),
				lipgloss.NewStyle().Bold(false).Render("\t  also hit \":\" to just hold the logs (continue with ESC)"),
				lipgloss.NewStyle().Bold(false).Render("\t- hit \"f\" to focus on a specific stream (coming soon, I'm working on it)"),
				lipgloss.NewStyle().Bold(false).Render("\t- hit \"cmd+f\" to apply a filter on all streams (coming soon, I'm working on it)"),
			}, "\n"),
		)
)

type Model struct {
	width, height int
}

func New(width int, height int) *Model {
	return &Model{
		width:  width,
		height: height,
	}
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width - 2 // account for margin
		m.height = msg.Height
	}

	return m, nil
}

func (m *Model) View() string {

	maxWidth := max(
		lipgloss.Width(welcomeLogo),
		lipgloss.Width(welcomeUsage),
		lipgloss.Width(howToText),
	)

	welcome := lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.PlaceHorizontal(
			maxWidth,
			lipgloss.Center,
			welcomeLogo,
		),
		lipgloss.PlaceHorizontal(
			maxWidth,
			lipgloss.Left,
			welcomeUsage,
		),
		lipgloss.PlaceHorizontal(
			maxWidth,
			lipgloss.Left,
			howToText,
		),
	)

	return lipgloss.NewStyle().
		Render(
			lipgloss.Place(
				m.width, m.height,
				lipgloss.Center, lipgloss.Center,
				welcome,
			),
		)
}

func max(vs ...int) int {

	var high int
	for _, v := range vs {
		high = int(math.Max(float64(high), float64(v)))
	}

	return high
}
