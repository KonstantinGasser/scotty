package header

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	text          string
	width, height int
}

var (
	border = lipgloss.NewStyle().
		Height(1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("45")).
		Foreground(lipgloss.Color("45")).
		Bold(true)
)

func New(w, h int, text string) *Model {
	return &Model{
		text:   text,
		width:  w,
		height: h,
	}
}

func (h *Model) Init() tea.Cmd                           { return nil }
func (h *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return h, nil }

func (h *Model) View() string {
	return lipgloss.Style.Render(border.Width(h.width), h.text)
}
