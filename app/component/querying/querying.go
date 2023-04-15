package querying

import (
	"github.com/KonstantinGasser/scotty/app/styles"
	"github.com/KonstantinGasser/scotty/debug"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	notImplemeted = lipgloss.NewStyle().
		Bold(true).
		AlignVertical(lipgloss.Center).
		AlignHorizontal(lipgloss.Center).
		Render("working on it!\n\nQuerying logs is not yet implemented")
)

type Model struct {
	ready         bool
	width, height int
}

func New() *Model {
	return &Model{
		ready:  false,
		width:  0,
		height: 0,
	}
}

func (model Model) Init() tea.Cmd {
	return nil
}

func (model *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	var (
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if !model.ready {
			model.ready = true
		}

		model.width = styles.ContentWidth(msg.Width)
		model.height = styles.ContentHeght(msg.Height)
		debug.Print("[browsing] Full-Width: %d Full-Height: %d Width: %d - Height: %d\n", msg.Width, msg.Height, model.width, model.height)
	}
	return model, tea.Batch(cmds...)
}

func (model Model) View() string {
	return lipgloss.NewStyle().Render(
		lipgloss.Place(
			model.width, model.height,
			lipgloss.Center, lipgloss.Center,
			notImplemeted,
		),
	)
}
