package welcome

import (
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	logo = lipgloss.NewStyle().
		MarginBottom(2).
		Render(
			strings.Join([]string{
				lipgloss.NewStyle().Foreground(lipgloss.Color("43")).Render(
					"███████╗ ██████╗ ██████╗ ████████╗████████╗██╗   ██╗",
				),
				lipgloss.NewStyle().Foreground(lipgloss.Color("43")).Render(
					"██╔════╝██╔════╝██╔═══██╗╚══██╔══╝╚══██╔══╝╚██╗ ██╔╝",
				),
				lipgloss.NewStyle().Foreground(lipgloss.Color("43")).Render(
					"███████╗██║     ██║   ██║   ██║      ██║    ╚████╔╝ ",
				),
				lipgloss.NewStyle().Foreground(lipgloss.Color("43")).Render(
					"╚════██║██║     ██║   ██║   ██║      ██║     ╚██╔╝ ",
				),
				lipgloss.NewStyle().Foreground(lipgloss.Color("43")).Render(
					"███████║╚██████╗╚██████╔╝   ██║      ██║      ██║",
				),
				lipgloss.NewStyle().Foreground(lipgloss.Color("43")).Render(
					"╚══════╝ ╚═════╝ ╚═════╝    ╚═╝      ╚═╝      ╚═╝",
				),
			}, "\n"),
		)

	helpUsage      = strings.Join([]string{}, "\n")
	styleHelpUsage = lipgloss.NewStyle().
			Render(helpUsage)

	helpBindings = []string{
		"format|SPC f",
		"query|SPC fq",
		"docs|SPC d",
	}
	// styleHelpBindings = lipgloss.NewStyle().
	// 			Render(helpBindings)

	// welcomeUsage = lipgloss.NewStyle().
	// 		MarginBottom(2).
	// 		Render(
	// 		strings.Join([]string{
	// 			lipgloss.NewStyle().Bold(true).Underline(true).Render("beam logs:\n"),
	// 			"\tfrom stderr:" + lipgloss.NewStyle().Bold(false).Render("go run -race my/awesome/app.go 2>&1 | beam -label=navigation_service"),
	// 			"\tfrom stdout:" + lipgloss.NewStyle().Bold(false).Render("cat uss_enterprise_engine_logs.log | beam -label=engine_service"),
	// 		}, "\n"),
	// 	)
	//
	// howToText = lipgloss.NewStyle().
	// 		Render(
	// 		strings.Join([]string{
	// 			lipgloss.NewStyle().Bold(true).Underline(true).Render("tips and notes:\n"),
	// 			lipgloss.NewStyle().Bold(false).Render("\t- hit \":\" and type an index to format a specific line. \nUse k/j to format the previous or next log"),
	// 			lipgloss.NewStyle().Bold(false).Render("\t  also hit \":\" to just hold the logs (continue with q)"),
	// 			lipgloss.NewStyle().Bold(false).Render("\t- hit \"cmd+f\" and type a comma separated list of beams \nto highlight them (ctrl+f: beam-one,beam-two)"),
	// 			lipgloss.NewStyle().Bold(false).Render("\t  while in the filter mode you can add/remove individual\nbeams by user the prefix +/- followed by the beam"),
	// 		}, "\n"),
	// 	)

	border = lipgloss.NewStyle().
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		Foreground(lipgloss.Color("43"))
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

	hBindings := []string{}
	for _, h := range helpBindings {
		parts := strings.Split(h, "|")
		hBindings = append(hBindings,
			lipgloss.JoinHorizontal(lipgloss.Left,
				lipgloss.NewStyle().Width(int(m.width)/3).Render(parts[0]),
				lipgloss.NewStyle().Width(int(m.width)/3).Render(parts[1]),
			))
	}

	maxWidth := max(
		lipgloss.Width(logo),
		lipgloss.Width(strings.Join(hBindings, "\n")),
	)

	welcome := lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.PlaceHorizontal(
			maxWidth,
			lipgloss.Center,
			logo,
		),
		// lipgloss.PlaceHorizontal(
		// 	maxWidth,
		// 	lipgloss.Left,
		// 	styleHelpUsage,
		// ),
		lipgloss.PlaceHorizontal(
			maxWidth,
			lipgloss.Center,
			strings.Join(hBindings, "\n"),
		),
	)

	// maxHeight := lipgloss.Height(welcome)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
		border.
			Render(
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
