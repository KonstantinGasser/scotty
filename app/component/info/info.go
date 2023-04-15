package info

import (
	"github.com/KonstantinGasser/scotty/app/styles"
	"github.com/KonstantinGasser/scotty/debug"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	minHeight = 15
)

var (
	style = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder())
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

		model.width = styles.InfoWidth(msg.Width)
		model.height = minHeight
		debug.Print("[info] Full-Width: %d Full-Height: %d Width: %d - Height: %d\n", msg.Width, msg.Height, model.width, model.height)
	}
	return model, tea.Batch(cmds...)
}

func (model Model) View() string {
	return style.
		Width(model.width).Height(minHeight).
		Render("")
}

func max(upper int, compare int) int {
	if compare > upper {
		return upper
	}
	return compare
}
