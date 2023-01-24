package pager

import (
	"bytes"

	"github.com/KonstantinGasser/scotty/app/styles"
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

type subscriber struct {
	label string
	color lipgloss.Color
}

// Logger implements the tea.Model interface.
// Furthermore, Logger allows to tail logs.
// Logger does not not store the logs its only
// porose is it to display them.
type Logger struct {
	buffer ring.Buffer
	writer bytes.Buffer

	beams map[string]lipgloss.Color
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
		writer: bytes.Buffer{},

		beams:  map[string]lipgloss.Color{},
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
		switch msg.Type {
		case tea.MouseWheelUp:
			break
		case tea.MouseWheelDown:
			break
		}
		return pager, cmd
	case tea.WindowSizeMsg:
		pager.width = msg.Width - 1   // pls fix this to constant so I will continue to understand
		pager.height = msg.Height - 1 // by now I have already no plan why it needs to be one - only now 2 messed things up

		// update viewport width an height
		pager.view.Width = pager.width
		pager.view.Height = pager.height

	// event dispatched each time a new stream connects to
	// the multiplexer. on-event we need to update the footer
	// model with the new stream information as well as update
	// the loggers state. The logger keeps track of connected beams
	// however only cares about the color to use when rendering the logs.
	// Logger will ensure that the color for the printed logs of a stream
	// are matching the color information in the footer
	case plexer.Subscriber:

		// stream was connected prior as such a color does already exist
		// and we only need to tell the footer about it again
		if color, ok := pager.beams[string(msg)]; ok {
			pager.footer, _ = pager.footer.Update(subscriber{
				label: string(msg),
				color: color,
			})
			return pager, tea.Batch(cmds...)
		}

		color, _ := styles.RandColor()
		pager.beams[string(msg)] = color

		pager.footer, _ = pager.footer.Update(subscriber{
			label: string(msg),
			color: color,
		})

		return pager, tea.Batch(cmds...)

	// event dispatched by the multiplexer each time a client/stream
	// sends a log linen.
	// The logger needs to add the ansi color code stored for the stream
	// to the dispatched message before adding the data to the ring buffer.
	// Once added to the ring buffer the logger queries for the latest N
	// records (where N is equal to the height of the current viewport.Model)
	// and pass the string to the viewport.Model for rendering
	case plexer.Message:

		color := pager.beams[msg.Label]

		p := []byte(lipgloss.NewStyle().Foreground(color).Render("[" + msg.Label + "] "))
		pager.buffer.Append(append(p, msg.Data...))

		err := pager.buffer.Window(
			&pager.writer,
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
	}

	// propagate events to child models.
	// in certain cases there will be an early return
	// in any of the cases above either because the event
	// is not relevant for any downstream model or because
	// ??? there was some other reason...??
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
