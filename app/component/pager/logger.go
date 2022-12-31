package pager

import (
	"github.com/KonstantinGasser/scotty/app/styles"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	marginLeft   = 2
	marginRight  = 2
	marginTop    = 1
	marginBottom = 1
)

var (
	pagerStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
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
	// available tty width and height
	// updates if changes
	width, height int
}

func NewLogger(width, height int) *Logger {

	w, h := width-(marginLeft+marginRight), height-(marginTop+marginBottom)
	vp := viewport.New(w, h)

	return &Logger{
		vp:     vp,
		width:  w,
		height: h,
	}
}

func (log *Logger) Init() tea.Cmd {
	return nil
}

func (log *Logger) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	var (
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		log.width = msg.Width
		log.height = msg.Height
	}

	return log, tea.Batch(cmds...)
}

func (log *Logger) View() string {
	return pagerStyle.Render(
		log.vp.View(),
	)
}
