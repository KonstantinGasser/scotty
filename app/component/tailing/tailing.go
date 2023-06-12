package tailing

import (
	"github.com/KonstantinGasser/scotty/app/bindings"
	"github.com/KonstantinGasser/scotty/app/styles"
	"github.com/KonstantinGasser/scotty/store"
	"github.com/KonstantinGasser/scotty/stream"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	borderMargin = 1

	unset = iota
	running
	paused
)

var (
	keyPause = key.NewBinding(
		key.WithKeys("p"),
	)

	keyScrollBottom = key.NewBinding(
		key.WithKeys("g"),
	)
)

type Model struct {
	ready         bool
	width, height int
	pager         store.Pager
	state         int
	bindings      *bindings.Map
}

func New(pager store.Pager) *Model {
	model := &Model{
		ready:    false,
		pager:    pager,
		state:    unset,
		bindings: bindings.NewMap(),
	}

	model.bindings.Bind("p").Action(func(msg tea.KeyMsg) tea.Cmd {
		if model.state == paused {
			model.state = running
			model.pager.Refresh()
			return RequestResume()
		}

		model.state = paused
		return RequestPause()
	})

	model.bindings.Bind("g").Action(func(msg tea.KeyMsg) tea.Cmd {
		model.pager.Refresh()
		return nil
	})

	model.bindings.Debug()

	return model
}

func (model *Model) Init() tea.Cmd {
	return nil
}

func (model *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case styles.Dimensions:
		model.setDimensions(msg.Width(), msg.Height())
		if !model.ready {
			model.pager.Reset(model.width, uint8(model.height))
			model.ready = true
			model.state = running
			break
		}

		model.pager.Rerender(model.width, model.height)

	case tea.KeyMsg:
		if !model.bindings.Matches(msg) {
			return model, tea.Batch(cmds...)
		}

		cmds = append(cmds, model.bindings.Exec(msg).Call(msg))
	case stream.Message:
		model.pager.MoveDown(model.state == paused)
	}

	return model, tea.Batch(cmds...)
}

func (model *Model) View() string {
	return model.pager.String()
}

func (model *Model) setDimensions(width, height int) {
	model.width = width
	model.height = height
}
