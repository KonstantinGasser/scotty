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

	between := int(float64(m.width) * 0.35)
	leaderBindings := lipgloss.JoinVertical(lipgloss.Center,
		lipgloss.PlaceHorizontal(between, lipgloss.Center, lipgloss.NewStyle().Render("Leader keys")),
		styles.SpaceBetween(between, leaderKeyActions, leaderKeyBindings, "."),
	)

	followBindings := lipgloss.JoinVertical(lipgloss.Center,
		lipgloss.PlaceHorizontal(between, lipgloss.Center, lipgloss.NewStyle().Render("Follow tab")),
		styles.SpaceBetween(between, followKeyActions, followKeyBindings, "."),
	)

	browseBindings := lipgloss.JoinVertical(lipgloss.Center,
		lipgloss.PlaceHorizontal(between, lipgloss.Center, lipgloss.NewStyle().Render("Browse tab")),
		styles.SpaceBetween(between, browseKeyActions, browseKeyBindings, "."),
	)
	maxWidth := max(
		lipgloss.Width(logo),
		lipgloss.Width(leaderBindings),
		lipgloss.Width(followBindings),
		lipgloss.Width(browseBindings),
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
