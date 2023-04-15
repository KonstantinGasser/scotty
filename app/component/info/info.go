package info

import (
	"fmt"

	"github.com/KonstantinGasser/scotty/app/event"
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
			break
		}
		model.beams[msg.label] = &msg
		model.ordered = append(model.ordered, msg.label)
	case event.Increment:
		debug.Print("[info] lookup: %s -> %v\n", string(msg), model.beams[string(msg)])
		if beam, ok := model.beams[string(msg)]; ok {
			debug.Print("[info] increment: %q\n", string(msg))
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
