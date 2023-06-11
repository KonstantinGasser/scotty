package tailing

import (
	"github.com/KonstantinGasser/scotty/app/bindings"
	"github.com/KonstantinGasser/scotty/app/styles"
	"github.com/KonstantinGasser/scotty/debug"
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
	bindings      bindings.Mapper
}

func New(binds bindings.Mapper, pager store.Pager) *Model {
	model := &Model{
		ready:    false,
		pager:    pager,
		state:    unset,
		bindings: binds,
	}

	binds.Map(keyPause, func(msg tea.KeyMsg) tea.Cmd {
		if model.state == paused {
			model.state = running
			model.pager.Refresh()
			return RequestResume()
		}

		model.state = paused
		return RequestPause()
	})

	binds.Map(keyScrollBottom, func(msg tea.KeyMsg) tea.Cmd {
		model.pager.Refresh()
		return nil
	})

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

	// case tea.KeyMsg:
	// 	switch {
	// 	case key.Matches(msg, model.bindings.Pause):
	// 		if model.state == paused {
	// 			model.state = running
	// 			model.pager.Refresh()
	// 			cmds = append(cmds, RequestResume())
	// 			break
	// 		}
	// 		model.state = paused
	// 		cmds = append(cmds, RequestPause())
	// 	// reset/reload pager with latest page
	// 	case key.Matches(msg, model.bindings.FastForward):
	// 		model.pager.GoToBottom()
	// 	}
	case styles.Dimensions:
		model.setDimensions(msg.Width(), msg.Height())
		debug.Print("[Tailing] Dimensions: width: %d - Height: %d\n", msg.Width(), msg.Height())
		if !model.ready {
			model.pager.Reset(model.width, uint8(model.height))
			model.ready = true
			model.state = running
			break
		}

		model.pager.Rerender(model.width, model.height)

	case stream.Message:
		model.pager.MoveDown(model.state == paused)
	case forceRefresh:
		model.pager.Refresh()
		return model, nil

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
