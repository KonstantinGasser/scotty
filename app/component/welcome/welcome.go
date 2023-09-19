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

	leaderKeyActions = []string{
		"follow logs",
		"browse logs",
		"query  logs",
	}

	leaderKeyBindings = []string{
		styles.Bold.Render("SPC f"),
		styles.Bold.Render("SPC b"),
		styles.Bold.Render("SPC s"),
	}

	followKeyActions = []string{
		"pause/continue",
		"scroll down",
	}

	followKeyBindings = []string{
		styles.Bold.Render("p"),
		styles.Bold.Render("g"),
	}

	browseKeyActions = []string{
		"select index",
		"next line",
		"previous line",
		"reload buffer",
	}

	browseKeyBindings = []string{
		styles.Bold.Render(":"),
		styles.Bold.Render("j"),
		styles.Bold.Render("k"),
		styles.Bold.Render("r"),
	}

	useageFormats = []string{
		"from stderr",
		"from stdout",
	}

	useageCmds = []string{
		lipgloss.NewStyle().Foreground(lipgloss.Color("#62fcaf")).Render("go run -race my/awesome/app.go 2>&1 | beam navigation_service"),
		lipgloss.NewStyle().Foreground(lipgloss.Color("#62fcaf")).Render("cat uss_enterprise_engine_logs.log | beam -d engine_service"),
	}
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

	betweenSmall := int(float64(m.width) * 0.35)
	betweenMedium := int(float64(m.width) * 0.4)

	leaderBindings := lipgloss.JoinVertical(lipgloss.Center,
		lipgloss.PlaceHorizontal(betweenSmall, lipgloss.Center, lipgloss.NewStyle().Render("Leader keys")),
		styles.SpaceBetween(betweenSmall, leaderKeyActions, leaderKeyBindings, "."),
	)

	followBindings := lipgloss.JoinVertical(lipgloss.Center,
		lipgloss.PlaceHorizontal(betweenSmall, lipgloss.Center, lipgloss.NewStyle().Render("Follow tab")),
		styles.SpaceBetween(betweenSmall, followKeyActions, followKeyBindings, "."),
	)

	browseBindings := lipgloss.JoinVertical(lipgloss.Center,
		lipgloss.PlaceHorizontal(betweenSmall, lipgloss.Center, lipgloss.NewStyle().Render("Browse tab")),
		styles.SpaceBetween(betweenSmall, browseKeyActions, browseKeyBindings, "."),
	)

	usageStderr := styles.SpaceBetween(betweenMedium, useageFormats[0:1], []string{".", "."}, ".")
	usageStdout := styles.SpaceBetween(betweenMedium, useageFormats[1:], []string{".", "."}, ".")

	maxWidth := max(
		lipgloss.Width(logo),
		lipgloss.Width(leaderBindings),
		lipgloss.Width(followBindings),
		lipgloss.Width(browseBindings),
		lipgloss.Width(usageStderr),
		lipgloss.Width(usageStdout),
	)

	welcome := lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.PlaceHorizontal(
			maxWidth,
			lipgloss.Center,
			logo,
		),
		lipgloss.PlaceHorizontal(
			maxWidth,
			lipgloss.Center,
			leaderBindings,
		),
		"\n",
		lipgloss.PlaceHorizontal(
			maxWidth,
			lipgloss.Center,
			followBindings,
		),
		"\n",
		lipgloss.PlaceHorizontal(
			maxWidth,
			lipgloss.Center,
			browseBindings,
		),
		"\n",
		lipgloss.PlaceHorizontal(
			maxWidth,
			lipgloss.Center,
			lipgloss.JoinVertical(lipgloss.Left,
				lipgloss.PlaceHorizontal(betweenMedium, lipgloss.Center, lipgloss.NewStyle().Render("Usage")),
				usageStderr,
				styles.FloatRight(betweenMedium, useageCmds[0]),
				usageStdout,
				styles.FloatRight(betweenMedium, useageCmds[1]),
			),
		),
	)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, welcome)

}

func max(vs ...int) int {

	var high int
	for _, v := range vs {
		high = int(math.Max(float64(high), float64(v)))
	}

	return high
}
