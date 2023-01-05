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

	beamSpacer = styles.Spacer(1).Render("")
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
	// slice of beams which are currently connected to scotty
	connectedBeams []string
}

func New(w, h int) *Model {
	return &Model{
		width:          w,
		height:         h,
		err:            nil,
		logCount:       0,
		connectedBeams: nil,
	}
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	var (
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width - 2 // account for margin
		m.height = msg.Height
		return m, nil
	case plexer.BeamNew:
		m.connectedBeams = append(m.connectedBeams, styles.StatusBarBeamInfo.Background(styles.RandColor()).Render((string(msg))))
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

	logText := "no logs beamed yet"
	if m.logCount > 0 {
		logText = "beamed logs: " + fmt.Sprint(m.logCount)
	}

	var items = []string{
		styles.StatusBarLogCount(logText),
		styles.Spacer(5).Render(""),
	}

	// add a little space between beam labels
	for i, beam := range m.connectedBeams {
		if i < len(m.connectedBeams)-1 {
			items = append(items, beamSpacer, beam)
			continue
		}
		// not space after last one thou
		items = append(items, beam)
	}

	if m.err != nil {
		items = append(items,
			styles.Spacer(2).Render(""), // add some space next to the beams
			styles.ErrorInfo(m.err.Error()),
		)
	}

	return footerStyle.
		Render(
			lipgloss.JoinHorizontal(lipgloss.Left,
				items...,
			),
		)
}
