package footer

import (
	"fmt"

	"github.com/KonstantinGasser/scotty/app/styles"
	plexer "github.com/KonstantinGasser/scotty/multiplexer"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	footerStyle = lipgloss.NewStyle().
		Margin(0, 2)
)

type Model struct {
	width, height int

	// any error happing anywhere
	// in the application should be shown
	// in the footer.
	// err represents the latest error
	err error
	// number of logs stream by all streams
	// dropping a stream results in logCount - len(stream)
	logCount int
}

func New(w, h int) *Model {
	return &Model{
		width:    w,
		height:   h,
		err:      nil,
		logCount: 0,
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

	switch msg := msg.(type) {
	case plexer.BeamError:
		// QUESTION @KonstantinGasser:
		// How do I unset the error say after 15 seconds?
		m.err = msg
	case plexer.BeamMessage:
		m.logCount++
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) View() string {
	if m.err != nil {
		return footerStyle.Render(
			lipgloss.JoinHorizontal(lipgloss.Left,
				styles.StatusBarLogCount("beamed logs: 4092"),
				styles.StatusBarBeamInfo("app_1"),
				styles.StatusBarBeamInfo("app_2"),
				styles.StatusBarBeamInfo("app_N"),
				styles.ErrorInfo(m.err.Error()),
			),
		)
	}
	return footerStyle.Render(
		lipgloss.JoinHorizontal(lipgloss.Left,
			styles.StatusBarLogCount(fmt.Sprintf("beamed logs: %d", m.logCount)),
			styles.StatusBarBeamInfo("app_1"),
			styles.StatusBarBeamInfo("app_2"),
			styles.StatusBarBeamInfo("app_N"),
		),
	)
}
