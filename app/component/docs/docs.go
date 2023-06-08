package docs

import (
	_ "embed"
	"fmt"

	"github.com/KonstantinGasser/scotty/app/styles"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
)

//go:embed DOCUMENTATION.md
var readme string

type Model struct {
	ready  bool
	width  int
	height int
	view   viewport.Model
}

func New() *Model {

	return &Model{
		ready: false,
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
	case styles.Content:
		model.width = msg.Width()
		model.height = msg.Height()

		if !model.ready {
			model.ready = true

			vp := viewport.New(model.width, model.height)
			model.view = vp

			out, err := glamour.Render(readme, "dark")
			if err != nil {
				model.view.SetContent(fmt.Errorf("unable to render documents...\n[ERROR]: %s", err).Error())
			} else {
				model.view.SetContent(out)
			}

		}

	}

	model.view, cmd = model.view.Update(msg)
	cmds = append(cmds, cmd)

	return model, tea.Batch(cmds...)
}

func (model *Model) View() string {
	return model.view.View()
}
