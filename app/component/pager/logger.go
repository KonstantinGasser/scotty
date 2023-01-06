package pager

import (
	"strings"

	"github.com/KonstantinGasser/scotty/app/styles"
	plexer "github.com/KonstantinGasser/scotty/multiplexer"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	marginLeft   = 0
	marginRight  = 10
	marginTop    = 30
	marginBottom = 0
)

var (
	pagerStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Margin(1, 1).
			Padding(1).
			BorderForeground(
			styles.ColorBorder,
		)

	pagerStyleActive = pagerStyle.Copy().
				UnsetBorderForeground().
				BorderForeground(
			styles.ColorBorderActive,
		)
)

// Logger implements the tea.Model interface.
// Furthermore, Logger allows to tail logs.
// Logger does not not store the logs its only
// porose is it to display them.
type Logger struct {

	// underlying model which handles
	// scrolling and rendering of the logs
	vp viewport.Model

	// serialized is a slice where each value is
	// the string representation of a log.
	// Values in this slice will be shown in the terminal
	// like tail -f would do
	//
	// THOUGHT @KonstantinGasser:
	// what should happen when a stream disconnects in case you restart a service?
	// These logs would need to be deleted
	// -> Refers to the overall question of how to store the logs; define requirements first pls
	serialized []string

	// describes any space in the Y-Axes which must be subtracted
	// from the height - when the terminal is resized we cannot simply
	// take the tea.WindowSizeMsg.Height but need to account for the offset
	offsetY int
	// available tty width and height
	// updates if changes
	width, height int
}

func NewLogger(width, height, offsetY int) *Logger {

	w, h := width, height-offsetY

	vp := viewport.New(w, h)
	vp.Height = h
	// no header we can render content in the first row
	vp.YPosition = 0
	// vp.HighPerformanceRendering = true
	vp.MouseWheelEnabled = true

	// vp.YPosition = 1
	return &Logger{
		vp:      vp,
		offsetY: offsetY,
		width:   w,
		height:  h,
	}
}

func (log *Logger) Init() tea.Cmd {
	return nil
}

func (log *Logger) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	var (
		cmds []tea.Cmd
		cmd  tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.MouseMsg:
		log.vp, cmd = log.vp.Update(msg)
		return log, cmd
	case tea.WindowSizeMsg:
		log.width = msg.Width
		log.height = msg.Height - log.offsetY

		// update viewport width an height
		log.vp.Width = log.width
		log.vp.Height = log.height

		// cmds = append(cmds, tea.SyncScrollArea(log.serialized, 0, log.height))
		return log, tea.Batch(cmds...)

	case plexer.BeamMessage:

		log.serialized = append(log.serialized, string(msg.Data))
		log.vp.SetContent(strings.Join(log.serialized, ""))

		log.vp.LineDown(1)
		// cmds = append(cmds, cmd, viewport.Sync(log.vp))

		return log, tea.Batch(cmds...)
	}

	log.vp, cmd = log.vp.Update(msg)
	cmds = append(cmds, cmd)

	return log, tea.Batch(cmds...)
}

func (log *Logger) View() string {
	return log.vp.View()
}
