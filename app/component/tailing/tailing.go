package tailing

import (
	"github.com/KonstantinGasser/scotty/app/event"
	"github.com/KonstantinGasser/scotty/app/styles"
	"github.com/KonstantinGasser/scotty/debug"
	"github.com/KonstantinGasser/scotty/multiplexer"
	"github.com/KonstantinGasser/scotty/store"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	borderMargin = 1
)

type Model struct {
	ready         bool
	width, height int
	pager         store.Pager
	view          viewport.Model
}

func New(pager store.Pager) *Model {
	return &Model{
		ready: false,
		pager: pager,
		view:  viewport.Model{},
	}
}

func (model *Model) Init() tea.Cmd {
	return nil
}

func (model *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmds []tea.Cmd
		cmd  tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if !model.ready {
			model.width = msg.Width - borderMargin
			model.height = styles.AvailableHeight(msg.Height)

			model.view = viewport.New(model.width, model.height)
			model.view.Width = model.width - borderMargin
			model.view.Height = model.height
			model.view.MouseWheelEnabled = true

			model.pager.Rerender(model.width, model.height)
			model.ready = true
			break
		}

		model.setDimensions(
			msg.Width,
			msg.Height,
		)
		model.pager.Rerender(model.width, model.height)

	case event.FormatInit:
		debug.Print("not implemented yet!")

	case event.FormatNext:
		debug.Print("not implemented yet!")

	case event.FormatPrevious:
		debug.Print("not implemented yet!")

	case multiplexer.Message:
		model.pager.MoveDown()
	}

	model.view, cmd = model.view.Update(msg)
	cmds = append(cmds, cmd)

	return model, tea.Batch(cmds...)
}

func (model *Model) View() string {
	return model.pager.String()
}

func (model *Model) setDimensions(width, height int) {
	model.width, model.height = width-borderMargin, styles.AvailableHeight(height)
	model.view.Width, model.view.Height = model.width, model.height
}
