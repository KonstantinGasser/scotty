package pager

import (
	"strings"

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
		Margin(0, 1).
		Padding(1)
)

type LogTailer interface {
	Tail(start int, end int) string
}

// Logger implements the tea.Model interface.
// Furthermore, Logger allows to tail logs.
// Logger does not not store the logs its only
// porose is it to display them.
type Logger struct {

	// underlying model which handles
	// scrolling and rendering of the logs
	view viewport.Model

	// store allows to retrieve logs ranging
	// from [start, end)
	store LogTailer

	// describes any space in the Y-Axes which must be subtracted
	// from the height - when the terminal is resized we cannot simply
	// take the tea.WindowSizeMsg.Height but need to account for the offset
	offsetY int
	// available tty width and height
	// updates if changes
	width, height int
}

func NewLogger(width, height, offsetY int) *Logger {

	w, h := width-1, height-offsetY-1 // -1 to margin top for testing

	view := viewport.New(w, h)
	view.Height = h
	// no header we can render content in the first row
	// view.HighPerformanceRendering = true
	view.MouseWheelEnabled = true

	// view.YPosition = 1
	return &Logger{
		view:    view,
		offsetY: offsetY,
		width:   w,
		height:  h,
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
		pager.width = msg.Width - 1
		pager.height = msg.Height - pager.offsetY - 1

		// update viewport width an height
		pager.view.Width = pager.width
		pager.view.Height = pager.height

		// cmds = append(cmds, tea.SyncScrollArea(pager.serialized, 0, pager.height))
		return pager, tea.Batch(cmds...)

	case plexer.BeamMessage:
		// eventually we will do no data processing
		// but only consume logs from the "store".
		// The "store" should offer an API to retrieve
		// [N,M) logs such that the pager.Logger only ever
		// has to render the current viewport logs.
		// Similar to a sliding window N,M will be shifted up/down
		// by the mouse wheel delta, however N-M stays constant unless
		// the height is changed by tea.WindowSizeMsg

		pager.view.SetContent(pager.store.Tail(0, 0))

		pager.view.LineDown(strings.Count("\n", string(msg.Data)))

		return pager, tea.Batch(cmds...)
	}

	pager.view, cmd = pager.view.Update(msg)
	cmds = append(cmds, cmd)

	return pager, tea.Batch(cmds...)
}

func (pager *Logger) View() string {
	return pagerStyle.Render(
		pager.view.View(),
	)
}
