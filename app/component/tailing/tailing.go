package tailing

import (
	"github.com/KonstantinGasser/scotty/app/styles"
	"github.com/KonstantinGasser/scotty/debug"
	"github.com/KonstantinGasser/scotty/multiplexer"
	"github.com/KonstantinGasser/scotty/store"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	borderMargin = 1

	unset = iota
	running
	paused
)

type Model struct {
	ready         bool
	width, height int
	pager         store.Pager
	state         int
	bindings      bindings
}

func New(pager store.Pager) *Model {
	return &Model{
		ready:    false,
		pager:    pager,
		state:    unset,
		bindings: defaultBindings,
	}
}

func (model *Model) Init() tea.Cmd {
	return nil
}

func (model *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, model.bindings.Pause):
			if model.state == paused {
				model.state = running
				break
			}
			model.state = paused
		}
	case tea.WindowSizeMsg:
		model.setDimensions(msg.Width, msg.Height)

		if !model.ready {
			model.pager.Reset(model.width, uint8(model.height))
			model.ready = true
			model.state = running
			break
		}

		model.pager.Rerender(model.width, model.height)

	case multiplexer.Message:
		if model.state == paused {
			break
		}
		model.pager.MoveDown()
	}

	return model, tea.Batch(cmds...)
}

func (model *Model) View() string {
	return model.pager.String()
}

func (model *Model) setDimensions(width, height int) {
	model.width = styles.ContentWidth(width)
	model.height = styles.ContentHeght(height)
}
