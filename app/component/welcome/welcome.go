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
				lipgloss.NewStyle().Bold(true).Render("usage:\n"),
				"\tfrom stderr: " + lipgloss.NewStyle().Bold(true).Render("go run -race my/awesome/app.go 2>&1 | beam -label=navigation_service"),
				"\tfrom stdout: " + lipgloss.NewStyle().Bold(true).Render("cat uss_enterprise_engine_logs.log | beam -label=engine_service"),
			}, "\n"),
		)

	welcomeQueries = lipgloss.NewStyle().
			Render(
			strings.Join([]string{
				lipgloss.NewStyle().Bold(true).Render("queries:\n"),
				"\tfilter stream(s): " + lipgloss.NewStyle().Bold(true).Render("filter beam=app_1 tracing_span='1e4851b8fe64ec763ad0'"),
				"\tapply statistics: " + lipgloss.NewStyle().Bold(true).Render("filter level=debug\n\t\t\t  | stats sum(tree_traversed)"),
				"\ttail -f a query : " + lipgloss.NewStyle().Bold(true).Render("tail |\n\t\t\t  filter level=debug\n\t\t\t  | stats sum(tree_traversed)"),
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
		lipgloss.Width(welcomeQueries),
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
			welcomeQueries,
		),
	)

	return lipgloss.NewStyle().
		Height(m.height).
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
