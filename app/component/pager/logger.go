package pager

import (
	"strings"

	"github.com/KonstantinGasser/scotty/debug"
	plexer "github.com/KonstantinGasser/scotty/multiplexer"
	"github.com/KonstantinGasser/scotty/ring"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	marginLeft   = 0
	marginRight  = 10
	marginTop    = 30
	marginBottom = 0

	footerHeight = 3
)

var (
	pagerStyle = lipgloss.NewStyle().
		Margin(0, 1).
		Padding(1)
)

// Logger implements the tea.Model interface.
// Furthermore, Logger allows to tail logs.
// Logger does not not store the logs its only
// porose is it to display them.
type Logger struct {
	buffer *ring.Buffer
	writer *strings.Builder
	// underlying model which handles
	// scrolling and rendering of the logs
	view viewport.Model

	// available tty width and height
	// updates if changes
	width, height int

	footer tea.Model
}

func NewLogger(width, height int) *Logger {

	w, h := width-1, height-footerHeight // -1 to margin top for testing

	view := viewport.New(w, h)
	view.Height = h
	// no header we can render content in the first row
	// view.HighPerformanceRendering = true
	view.MouseWheelEnabled = true

	// view.YPosition = 1
	return &Logger{
		buffer: ring.New(uint32(12)),
		writer: &strings.Builder{},
		view:   view,
		width:  w,
		height: h,
		footer: newFooter(w, h),
	}
}

func (pager *Logger) Init() tea.Cmd {
	return nil
}

func (pager *Logger) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	var (
		cmds []tea.Cmd
		cmd  tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.MouseMsg:
		pager.view, cmd = pager.view.Update(msg)
		return pager, cmd
	case tea.WindowSizeMsg:
		pager.width = msg.Width - 1   // pls fix this to constant so I will continue to understand
		pager.height = msg.Height - 1 // by now I have already no plan why it needs to be one - only now 2 messed things up

		// update viewport width an height
		pager.view.Width = pager.width
		pager.view.Height = pager.height

	case plexer.BeamMessage:

		p := []byte("[" + msg.Label + "] ")
		pager.buffer.Append(append(p, msg.Data...))

		err := pager.buffer.Window(
			pager.writer,
			pager.height,
			nil,
		)
		if err != nil {
			debug.Debug(err.Error())
		}
		pager.view.SetContent(pager.writer.String())
		pager.writer.Reset()

		// this has one flaw; if a log with longer then the width of the terminal it will be wrapped -> >1 line
		pager.view.LineDown(1)

		return pager, tea.Batch(cmds...)
	}

	pager.view, cmd = pager.view.Update(msg)
	cmds = append(cmds, cmd)

	pager.footer, cmd = pager.footer.Update(msg)
	cmds = append(cmds, cmd)

	return pager, tea.Batch(cmds...)
}

func (pager *Logger) View() string {
	return lipgloss.JoinVertical(lipgloss.Left,
		pagerStyle.Render(
			pager.view.View(),
		),
		pager.footer.View(),
	)
}
