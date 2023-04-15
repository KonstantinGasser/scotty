package info

import (
	"fmt"

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

const (
	connected = iota
	disconnected
)

type beam struct {
	label string
	count int
	state int
}

type Model struct {
	ready         bool
	width, height int
	beams         []beam
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
		model.height = styles.InfoHeight(msg.Height)
		debug.Print("[info] Full-Width: %d Full-Height: %d Width: %d - Height: %d\n", msg.Width, msg.Height, model.width, model.height)
	case beam:
		model.beams = append(model.beams, msg)
	}
	return model, tea.Batch(cmds...)
}

func (model Model) View() string {

	var beams = []string{}
	for i, beam := range model.beams {
		if i < len(model.beams)-1 {
			beams = append(beams, beam.label+": "+fmt.Sprint(beam.count), " - ")
			continue
		}
		beams = append(beams, beam.label+": "+fmt.Sprint(beam.count))
	}
	debug.Print("%v\n", beams)
	return style.
		Width(model.width).Height(model.height).
		Render(
			lipgloss.JoinHorizontal(lipgloss.Top, beams...),
		)
}

func max(upper int, compare int) int {
	if compare > upper {
		return upper
	}
	return compare
}

func NewBeam(label string, color lipgloss.Color) tea.Msg {
	return beam{
		label: lipgloss.NewStyle().Foreground(color).Render(label),
		count: 0,
		state: connected,
	}
}
