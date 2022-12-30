package footer

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

func (h *Model) Init() tea.Cmd {
	return nil
}

func (h *Model) SetSize(width, height int) {
	h.width, h.height = width, height
}

func (h *Model) View() string {
	return ""
}
