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
)

const (
	bottomSectionHeight = 1

	// wow literally no idea why this number hence
	// the variable name - if you get why tell me and
	// pls open a PR..else pls don't change it
	magicNumber = 2
)

var (
	pagerStyle = lipgloss.NewStyle()
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

	// selected represents an index of the buffer
	// which was initially requested to be parsed.
	// It can be decremented or incremented to parse
	// the previous or next item in the buffer
	selected int
	// awaitInput indicated if ECS is pressed.
	// if awaitInput == false the input for commands
	// is focused else moved out of focus
	awaitInput bool
	footer     tea.Model
	cmd        tea.Model
}

func NewLogger(width, height int) *Logger {

	w, h := width, height-bottomSectionHeight-magicNumber // -1 to margin top for testing

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
		switch msg.String() {
		case ":":
			if pager.awaitInput {
				break
			}
			pager.awaitInput = true
			width := pager.width - int(pager.width/3) - (1 + 2)

			pager.setDimensions(
				width,
				pager.height,
			)

		case "esc":
			if !pager.awaitInput {
				break
			}

			width, height, err := styles.WindowSize()
			if err != nil {
				debug.Print("[tea.KeyMsg(esc)] unable to get tty width and height: %w\n", err)
			}

			pager.setDimensions(
				width,
				height-bottomSectionHeight-magicNumber,
			)

			pager.awaitInput = false

		// selects the previous log line to be parsed
		// and displayed. Input ignores when selected <= 0
		case "up", "k":
			if pager.selected <= 0 {
				break
			}
			pager.selected--

			parsed := pager.parse(pager.selected)
			pager.cmd, _ = pager.cmd.Update(
				parsed(),
			)
			return pager, nil

		// selects the next log line to be parsed and
		// displayed. Input ignored when selected >= buffer.cap
		case "down", "j":
			if pager.selected >= int(pager.buffer.Cap()) {
				break
			}
			pager.selected++

			parsed := pager.parse(pager.selected)
			pager.cmd, _ = pager.cmd.Update(
				parsed(),
			)
			return pager, nil
		}

	// event dispatched from bubbletea when the screen size changes.
	// We need to update the pager and pager.view width and height.
	// However, if the parsing mode is on the width is only 2/3
	// of the available screen size.
	case tea.WindowSizeMsg:

		width := msg.Width
		height := msg.Height - bottomSectionHeight - magicNumber

		if pager.awaitInput && pager.cmd != nil {
			width = msg.Width - int(msg.Width/3) - (1 + 2) // magic number + margin between logs and parsed code
		}

		pager.setDimensions(
			width,
			height,
		)

	// event dispatched each time a beam disconnects from scotty.
	// The message itself is the label of the stream which
	// disconnected. On a disconnect we need to recompute the
	// length of the longest stream label in order to maintain
	// pretty indention for logging the logs with the label prefix
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

		space := pager.maxLabelLength - len(msg.Label)
		if space < 0 {
			space = 0
		}

		prefix := msg.Label + strings.Repeat(" ", space) + " | "
		colored := []byte(lipgloss.NewStyle().Foreground(color).Render(prefix))
		pager.buffer.Append(append(colored, msg.Data...))

		err := pager.buffer.Window(
			&pager.writer,
			pager.height,
			ring.WithLineWrap(pager.width-1), // -1 as we need to account for the pixels reserved for the border
		)
		if err != nil {
			debug.Debug(err.Error())
		}

		pager.view.SetContent(pager.writer.String())
		pager.writer.Reset()

		// this has one flaw; if a log with longer then the width of the terminal it will be wrapped -> >1 line
		pager.view.GotoBottom()

	// event dispatched by the command model whenever the user
	// enters on an input requesting to parse a log line.
	// The msg of type parserIndex is an integer and represents
	// the captured requested index.
	case parserIndex:
		parsed := pager.parse(int(msg))
		pager.cmd, _ = pager.cmd.Update(
			parsed(),
		)
		return pager, nil
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

	pager.cmd, cmd = pager.cmd.Update(msg)
	cmds = append(cmds, cmd)

	return pager, tea.Batch(cmds...)
}

func (pager *Logger) View() string {

	if pager.awaitInput && pager.cmd != nil {
		return lipgloss.JoinVertical(lipgloss.Left,
			lipgloss.JoinHorizontal(lipgloss.Left,
				pagerStyle.
					Border(lipgloss.RoundedBorder()).
					BorderForeground(styles.ColorBorder).
					Width(pager.width).
					Height(pager.height).
					Render(
						pager.view.View(),
					),
				pager.cmd.View(),
			),
			pager.footer.View(),
		)
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		pagerStyle.
			Padding(1).
			Render(
				pager.view.View(),
			),
		pager.footer.View(),
	)
}

func (pager *Logger) parse(index int) tea.Cmd {
	parsed, err := pager.buffer.At(index, ring.WithIndentation())
	if err != nil {
		debug.Print("unable to parse buffer item at index=%d: %v\n", index, err)
	}
	return emitParsed(parsed)
}

func (pager *Logger) setDimensions(width, height int) {
	pager.width, pager.height = width, height
	pager.view.Width, pager.view.Height = width, height
}
