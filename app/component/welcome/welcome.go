package welcome

import (
	"math"
	"strings"

	"github.com/KonstantinGasser/scotty/app/styles"
	"github.com/charmbracelet/lipgloss"
)

var (
	logo = lipgloss.NewStyle().
		MarginBottom(2).
		Render(
			strings.Join([]string{
				lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0f7b")).Render(
					"███████╗ ██████╗ ██████╗ ████████╗████████╗██╗   ██╗",
				),
				lipgloss.NewStyle().Foreground(lipgloss.Color("#fd3863")).Render(
					"██╔════╝██╔════╝██╔═══██╗╚══██╔══╝╚══██╔══╝╚██╗ ██╔╝",
				),
				lipgloss.NewStyle().Foreground(lipgloss.Color("#fc5154")).Render(
					"███████╗██║     ██║   ██║   ██║      ██║    ╚████╔╝ ",
				),
				lipgloss.NewStyle().Foreground(lipgloss.Color("#fb6648")).Render(
					"╚════██║██║     ██║   ██║   ██║      ██║     ╚██╔╝ ",
				),
				lipgloss.NewStyle().Foreground(lipgloss.Color("#fa7f39")).Render(
					"███████║╚██████╗╚██████╔╝   ██║      ██║      ██║",
				),
				lipgloss.NewStyle().Foreground(lipgloss.Color("#f89b29")).Render(
					"╚══════╝ ╚═════╝ ╚═════╝    ╚═╝      ╚═╝      ╚═╝",
				),
			}, "\n"),
		)

	infoGlobal = lipgloss.NewStyle().Render(
		strings.Join([]string{
			lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#bdbdbe")).Underline(true).Render("Global keys"),
			styles.Bold.Render("SPC f") + " ● open view to follow/tail all logs",
			styles.Bold.Render("SPC b") + " ● open view to browse all logs",
			styles.Bold.Render("SPC s") + " ● open view to query the logs",
		}, "\n"),
	)

	infoTabFollow = lipgloss.NewStyle().Render(
		strings.Join([]string{
			lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#bdbdbe")).Underline(true).Render("\n[View] Following/Tailing"),
			styles.Bold.Render("p") + " ● to pause the tailing (if not paused). Use p to contiune tailing",
			styles.Bold.Render("g") + " ● to tail the latest logs. Useful while in paused state",
		}, "\n"),
	)

	infoTabBrowsing = lipgloss.NewStyle().Render(
		strings.Join([]string{
			lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#bdbdbe")).Underline(true).Render("\n[View] Browsing"),
			styles.Bold.Render(":") + " ● to enable prompt input for index selection",
			styles.Bold.Render("\tenter") + " ● sets the sected formatted log to the requested index",
			styles.Bold.Render("j") + " ● to format the next log line",
			styles.Bold.Render("k") + " ● to format the previous log line",
			styles.Bold.Render("r") + " ● to reload the formatter with the latest log lines",
		}, "\n"),
	)

	infoUsage = lipgloss.NewStyle().
			MarginBottom(2).
			Render(
			strings.Join([]string{
				lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#bdbdbe")).Underline(true).Render("\nUsage"),
				styles.Bold.Render("from stderr: ") + lipgloss.NewStyle().Render("go run -race my/awesome/app.go 2>&1 | beam navigation_service"),
				styles.Bold.Render("from stdout: ") + lipgloss.NewStyle().Render("cat uss_enterprise_engine_logs.log | beam -d engine_service"),
			}, "\n"),
		)
)

type Model struct {
	ready         bool
	width, height int
}

func New(width int, height int) *Model {
	return &Model{
		ready:  false,
		width:  width,
		height: height,
	}
}

func (m *Model) View() string {

	maxWidth := max(
		lipgloss.Width(logo),
		lipgloss.Width(infoGlobal),
		lipgloss.Width(infoTabFollow),
		lipgloss.Width(infoTabBrowsing),
		lipgloss.Width(infoUsage),
	)

	welcome := lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.PlaceHorizontal(
			maxWidth,
			lipgloss.Center,
			logo,
		),
		lipgloss.PlaceHorizontal(
			maxWidth,
			lipgloss.Left,
			infoGlobal,
		),
		lipgloss.PlaceHorizontal(
			maxWidth,
			lipgloss.Left,
			infoTabFollow,
		),
		lipgloss.PlaceHorizontal(
			maxWidth,
			lipgloss.Left,
			infoTabBrowsing,
		),
		lipgloss.PlaceHorizontal(
			maxWidth,
			lipgloss.Left,
			infoUsage,
		),
	)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
		welcome,
	)

}

func max(vs ...int) int {

	var high int
	for _, v := range vs {
		high = int(math.Max(float64(high), float64(v)))
	}

	return high
}
