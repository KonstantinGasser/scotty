package pager

import (
	"bytes"
	"strings"

	"github.com/KonstantinGasser/scotty/app/styles"
	"github.com/KonstantinGasser/scotty/debug"
	plexer "github.com/KonstantinGasser/scotty/multiplexer"
	"github.com/KonstantinGasser/scotty/ring"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wrap"
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
	// stores the length of the longest stream
	// label in order to align the start of the logs
	maxLabelLength int
	// underlying model which handles
	// scrolling and rendering of the logs
	view viewport.Model

	// available tty width and height
	// updates if changes
	width, height int

	// awaitInput indicated if ECS is pressed.
	// if awaitInput == false the input for commands
	// is focused else moved out of focus
	awaitInput bool
	footer     tea.Model
	cmd        tea.Model
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

		beams:          map[string]lipgloss.Color{},
		maxLabelLength: 0,
		view:           view,
		width:          w,
		height:         h,
		awaitInput:     false,
		footer:         newFooter(w, h),
		cmd:            newCommand(w, h),
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
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			pager.awaitInput = !pager.awaitInput
		}
	case tea.WindowSizeMsg:
		pager.width = msg.Width - 1   // pls fix this to constant so I will continue to understand
		pager.height = msg.Height - 1 // by now I have already no plan why it needs to be one - only now 2 messed things up

		// update viewport width an height
		pager.view.Width = pager.width
		pager.view.Height = pager.height

	case plexer.Unsubscribe:
		// we only need to reassign the max value
		// if the current max is disconnecting
		if len(msg) >= pager.maxLabelLength {
			max := 0
			for label := range pager.beams {
				if len(label) > max && label != string(msg) {
					max = len(label)
				}
			}
			pager.maxLabelLength = max
		}

	// event dispatched each time a new stream connects to
	// the multiplexer. on-event we need to update the footer
	// model with the new stream information as well as update
	// the loggers state. The logger keeps track of connected beams
	// however only cares about the color to use when rendering the logs.
	// Logger will ensure that the color for the printed logs of a stream
	// are matching the color information in the footer
	case plexer.Subscriber:

		// update max label length for indenting
		// while displaying logs
		if len(msg) > pager.maxLabelLength {
			pager.maxLabelLength = len(msg)
		}

		label := string(msg)

		if _, ok := pager.beams[label]; !ok {
			color, _ := styles.RandColor()
			pager.beams[label] = color

			pager.footer, _ = pager.footer.Update(subscriber{
				label: label,
				color: color,
			})
		}

		pager.footer, _ = pager.footer.Update(subscriber{
			label: label,
			color: pager.beams[label],
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
		prefix := "[" + msg.Label + "]" + strings.Repeat(" ", pager.maxLabelLength-len(msg.Label))
		colored := []byte(lipgloss.NewStyle().Foreground(color).Render(prefix))
		pager.buffer.Append(append(colored, msg.Data...))

		err := pager.buffer.Window(
			&pager.writer,
			pager.height,
			WithLineWrap(pager.width-len(prefix)),
		)
		if err != nil {
			debug.Debug(err.Error())
		}
		pager.view.SetContent(pager.writer.String())
		pager.writer.Reset()

		// this has one flaw; if a log with longer then the width of the terminal it will be wrapped -> >1 line
		pager.view.GotoBottom()

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

	var bottom = ""
	if pager.awaitInput {
		bottom = pager.cmd.View()
	} else {
		bottom = pager.footer.View()
	}
	return lipgloss.JoinVertical(lipgloss.Left,
		pagerStyle.Render(
			pager.view.View(),
		),
		bottom,
	)
}

// WithMultipleLines adds \n in the slice of bytes such
// that the resulting slice of bytes respects the provided
// max width.
func WithLineWrap(width int) func([]byte) []byte {
	return func(b []byte) []byte {
		return wrap.Bytes(b, width)
	}
}
