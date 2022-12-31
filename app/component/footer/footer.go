package footer

import (
	"github.com/KonstantinGasser/scotty/app/styles"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	footerStyle = lipgloss.NewStyle().
		Margin(0, 2)
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
	return footerStyle.Render(
		lipgloss.JoinHorizontal(lipgloss.Left,
			styles.StatusBarLogCount("beamed logs: 4092"),
			styles.StatusBarBeamInfo("app_1"),
		),
	)
}
