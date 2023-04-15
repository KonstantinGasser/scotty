package info

import (
	"fmt"

	"github.com/KonstantinGasser/scotty/app/event"
	"github.com/KonstantinGasser/scotty/app/styles"
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
	label   string
	colored string
	count   int
	state   int
}

func (b *beam) increment() { b.count++ }

type Model struct {
	ready         bool
	width, height int
	beams         map[string]*beam
	ordered       []string
}

func New() *Model {
	return &Model{
		ready:   false,
		width:   0,
		height:  0,
		beams:   map[string]*beam{},
		ordered: []string{},
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
	case beam:
		if _, ok := model.beams[msg.label]; ok {
			model.beams[msg.label].state = connected
			break
		}
		model.beams[msg.label] = &msg
		model.ordered = append(model.ordered, msg.label)
	case DisconnectBeam:
		if _, ok := model.beams[string(msg)]; !ok {
			break
		}
		model.beams[string(msg)].state = disconnected

	case event.Increment:
		if beam, ok := model.beams[string(msg)]; ok {
			beam.increment()
		}
	}
	return model, tea.Batch(cmds...)
}

func (model Model) View() string {

	var beams = []string{}
	var status = "●"
	for i, label := range model.ordered {
		if model.beams[label].state == disconnected {
			status = "◌"
		}

		if i < len(model.ordered)-1 {
			beams = append(beams, status+" "+model.beams[label].colored+": "+fmt.Sprint(model.beams[label].count), " - ")
			continue
		}
		beams = append(beams, status+" "+model.beams[label].colored+": "+fmt.Sprint(model.beams[label].count))
	}

	return style.
		Padding(0, 1).
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
		label:   label,
		colored: lipgloss.NewStyle().Foreground(color).Render(label),
		count:   0,
		state:   connected,
	}
}

type DisconnectBeam string
