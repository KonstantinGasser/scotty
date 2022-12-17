package header

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	logo = strings.Join([]string{
		"█▀ █▀▀ █▀█ ▀█▀ ▀█▀ █▄█",
		"▄█ █▄▄ █▄█ ░█░ ░█░ ░█░",
	}, "\n")
)

type Model struct {
	text          string
	width, height int
}

var (
	border = lipgloss.NewStyle().
		Align(lipgloss.Left).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("135")).
		// Foreground(lipgloss.Color("45")).
		Bold(true)
)

func New(w, h int, text string) *Model {
	return &Model{
		text:   text,
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

	return border.Width(h.width).Render(
		lipgloss.JoinHorizontal(lipgloss.Left,
			lipgloss.NewStyle().
				Foreground(lipgloss.Color("135")).
				Render(logo),
		),
	)
	// return lipgloss.Style.Render(border.Width(h.width), h.text+"("+fmt.Sprint(h.width)+")")
}
