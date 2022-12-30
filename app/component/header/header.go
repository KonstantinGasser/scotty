package header

import (
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	width, height int
}

func New(w, h int) *Model {
	return &Model{
		width:  w,
		height: h,
	}
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	var (
		cmds []tea.Cmd
		// cmd  tea.Cmd
	)

	return m, tea.Batch(cmds...)
}

func (m *Model) View() string {
	return "I am the header"
}
