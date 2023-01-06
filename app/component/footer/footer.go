package footer

import (
	"fmt"
	"sync"

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

	// guards below fields
	mtx sync.RWMutex
	// number of logs stream by all streams
	// dropping a stream results in logCount - len(stream)
	logCount int
	// slice of beams which are currently connected to scotty
	connectedBeams map[string]int
}

func New(w, h int) *Model {
	return &Model{
		width:  w,
		height: h,
		err:    nil,

		mtx:            sync.RWMutex{},
		logCount:       0,
		connectedBeams: map[string]int{},
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
		m.mtx.RLock()
		if _, ok := m.connectedBeams[string(msg)]; ok {
			m.connectedBeams[string(msg)] = 0
			m.mtx.RUnlock()
			break
		}
		m.mtx.RLock()
		m.connectedBeams[string(msg)] = 0
		m.mtx.RUnlock()

	case plexer.BeamError:
		// QUESTION @KonstantinGasser:
		// How do I unset the error say after 15 seconds?
		m.err = msg
	case plexer.BeamMessage:
		// plexer.BeamMessage needs to be extended with
		// information about the stream such as the label of it
		// only then we can increase the respective count
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
	var i int
	for beam, count := range m.connectedBeams {
		if i < len(m.connectedBeams) {
			items = append(items, beamSpacer, beam+":"+fmt.Sprint(count))
			i++
			continue
		}
		// not space after last one thou
		items = append(items, beam)
		i++
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
