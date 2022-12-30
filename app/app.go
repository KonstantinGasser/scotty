package app

import (
	"fmt"
	"math"
	"strings"

	"github.com/KonstantinGasser/scotty/app/component/footer"
	"github.com/KonstantinGasser/scotty/app/component/header"
	"github.com/KonstantinGasser/scotty/app/styles"
	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	welcomeLogo = lipgloss.NewStyle().
			MarginBottom(4).
			Foreground(styles.ColorLogo).
			Render(
			strings.Join([]string{
				"____   ___  __  ____  ____  _  _ ",
				"/ ___) / __)/  \\(_  _)(_  _)( \\/ )",
				"\\___ \\( (__(  O ) )(    )(   )  / ",
				"(____/ \\___)\\__/ (__)  (__) (__/",
			}, "\n"),
		)

	welcomeUsage = lipgloss.NewStyle().
			Render(
			strings.Join([]string{
				"usage:\n",
				"\tfrom stderr: " + lipgloss.NewStyle().Bold(true).Render("go run -race my/awesome/app.go 2>&1 | beam"),
				"\tfrom stdout: " + lipgloss.NewStyle().Bold(true).Render("cat uss_enterprise_engine_logs.log | beam"),
			}, "\n"),
		)

	welcomeNews = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(styles.ColorBorder).
			Padding(1, 2).
			MarginBottom(2).
			Render(
			strings.Join([]string{
				"Changelog:\n",
				"🚀 query single log streams or all together",
				"\n",
				"🐞 fixed issue #24 scrolling for logs failed when logs exceed terminal width",
			}, "\n"),
		)
)

const (
	empty = iota
	logging
)

type App struct {

	// help view
	help help.Model

	keys bindings
	// header: shows logo
	header tea.Model

	// footer: displays stats (log/stream count)
	// has input field for query
	footer tea.Model

	// views is a map of tea.Model which can be either of type component/pager
	// or component/view and represent the views which should be displayed
	// in the next render
	// TODO: do we want to allow to show multiple views? or only one at the time?
	// background of the question is 1. rending of the string is more complex
	// 2. showing more then one view reduces the space a single view can take.
	// On the other hand in the future it would be cool if the user could store
	// "dashboard" configurations which can be used to pre-initialize scotty
	views map[int]tea.Model

	// width and height of the terminal
	// updates as it changes
	width, height int
	state         int
}

func New() (*App, error) {

	width, height, err := windowSize()
	if err != nil {
		return nil, fmt.Errorf("unable to determine the initial dimensions of the terminal: %w", err)
	}

	return &App{
		help:   help.New(),
		keys:   defaultBindings,
		header: header.New(width, height),
		views:  make(map[int]tea.Model),
		footer: footer.New(width, height),
		width:  width,
		height: height,
		state:  empty,
	}, nil
}

func (app *App) Init() tea.Cmd {
	return nil
}

func (app *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	var (
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if quite := app.resolveKey(msg); quite != nil {
			return app, quite
		}
	case tea.WindowSizeMsg:
		app.width = msg.Width
		app.height = msg.Height
	}

	return app, tea.Batch(cmds...)
}

func (app *App) View() string {

	if app.state == empty {

		maxWidth := max(
			lipgloss.Width(welcomeLogo),
			lipgloss.Width(welcomeUsage),
			lipgloss.Width(welcomeNews),
		)

		welcome := lipgloss.JoinVertical(lipgloss.Left,
			lipgloss.PlaceHorizontal(
				maxWidth,
				lipgloss.Center,
				welcomeLogo,
			),
			welcomeNews,
			lipgloss.PlaceHorizontal(
				maxWidth,
				lipgloss.Left,
				welcomeUsage,
			),
			app.footer.View(),
		)

		return lipgloss.JoinVertical(lipgloss.Left,
			lipgloss.NewStyle().
				Height(app.heightWithoutFooter()).
				Render(
					lipgloss.Place(
						app.width, app.height,
						lipgloss.Center, lipgloss.Center,
						welcome,
					),
				),
			app.footer.View(),
		)
	}
	return lipgloss.JoinVertical(lipgloss.Left,
		app.header.View(),
		app.footer.View(),
	)
}

func (app *App) heightWithoutFooter() int {
	return app.height - lipgloss.Height(app.footer.View())
}

func max(vs ...int) int {

	var high int
	for _, v := range vs {
		high = int(math.Max(float64(high), float64(v)))
	}

	return high
}
