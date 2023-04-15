package tailing

import (
	"github.com/KonstantinGasser/scotty/app/event"
	"github.com/KonstantinGasser/scotty/app/styles"
	"github.com/KonstantinGasser/scotty/debug"
	"github.com/KonstantinGasser/scotty/multiplexer"
	"github.com/KonstantinGasser/scotty/store"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	borderMargin = 1
)

type Model struct {
	ready         bool
	width, height int
	pager         store.Pager
}

func New(pager store.Pager) *Model {
	return &Model{
		ready: false,
		pager: pager,
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
	case event.DimensionMsg:
		model.setDimensions(
			msg.AvailableWidth,
			msg.AvailableHeight,
		)

		if !model.ready {
			model.pager.Reset(model.width, uint8(model.height))
			model.ready = true
		}

	case tea.WindowSizeMsg:
		model.setDimensions(msg.Width, msg.Height)

		if !model.ready {
			model.pager.Reset(model.width, uint8(model.height))
			model.ready = true
			break
		}

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

	return model, tea.Batch(cmds...)
}

func (model *Model) View() string {
	return model.pager.String()
}

func (model *Model) setDimensions(width, height int) {
	model.width = styles.ContentWidth(width)
	model.height = styles.ContentHeght(height)
}
