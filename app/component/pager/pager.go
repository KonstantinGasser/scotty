package pager

import (
	"bytes"
	"strings"

	"github.com/KonstantinGasser/scotty/app/styles"
	"github.com/KonstantinGasser/scotty/debug"
	plexer "github.com/KonstantinGasser/scotty/multiplexer"
	"github.com/KonstantinGasser/scotty/ring"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	bottomSectionHeight = 1
	inputSectionHeight  = 1

	borderMargin = 1

	// wow literally no idea why this number hence
	// the variable name - if you get why tell me and
	// pls open a PR..else pls don't change it
	magicNumber = 2

	awaitInput = iota + 1
	hasInput
)

var (
	pagerStyle = lipgloss.NewStyle().Padding(0, 1)

	inputStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(styles.DefaultColor.Border)
)

type subscriber struct {
	label string
	color lipgloss.Color
}

// Model implements the tea.Model interface.
// Furthermore, Model allows to tail logs.
// Model does not not store the logs its only
// porose is it to display them.
type Model struct {
	buffer *ring.Buffer
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

	// offsetStart if used when paging through the logs
	// and formatting log lines. It refers to the index
	// with which the pager starts (first log of the page)
	offsetStart int

	// pageSize refers to the number of items currently
	// visible in the view - line wraps are not included
	// an item which takes up two lines counts as one
	pageSize int
}

func New(width, height int, buffer *ring.Buffer) *Model {
	w, h := width-borderMargin, height-bottomSectionHeight-inputSectionHeight-magicNumber

	view := viewport.New(w, h)
	view.Height = h
	view.Width = w
	view.MouseWheelEnabled = true
	view.Style = pagerStyle.Width(w)

	debug.Print("[model.New] width: %d (%d), height: %d (%d)\n", w, view.Width, h, view.Height)

	input := textinput.New()
	input.Placeholder = "line number (use k/j to move and ESC/q to exit)"
	input.Prompt = ":"

	return &Model{
		buffer: buffer,
		writer: bytes.Buffer{},

		beams:          map[string]lipgloss.Color{},
		maxLabelLength: 0,
		view:           view,
		width:          w,
		height:         h,
	}
}

func (model *Model) Init() tea.Cmd {
	return nil
}

func (model *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	var (
		cmds []tea.Cmd
		cmd  tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.MouseMsg:
		model.view, cmd = model.view.Update(msg)
		cmds = append(cmds, cmd)
	case tea.KeyMsg:

	// event dispatched from bubbletea when the screen size changes.
	// We need to update the pager and model.view width and height.
	// However, if the parsing mode is on the width is only 2/3
	// of the available screen size.
	case tea.WindowSizeMsg:

		model.setDimensions(
			msg.Width,
			msg.Height,
		)

		contents, _ := model.peekBuffer(
			model.height,
			ring.WithLineWrap(model.width),
		)
		model.writer.Reset()
		model.view.SetContent(contents)
		model.view.GotoBottom()

	// event dispatched each time a beam disconnects from scotty.
	// The message itself is the label of the stream which
	// disconnected. On a disconnect we need to recompute the
	// length of the longest stream label in order to maintain
	// pretty indention for logging the logs with the label prefix
	case plexer.Unsubscribe:
		// // we only need to reassign the max value
		// // if the current max is disconnecting
		// if len(msg) >= model.maxLabelLength {
		// 	max := 0
		// 	for label := range model.beams {
		// 		if len(label) > max && label != string(msg) {
		// 			max = len(label)
		// 		}
		// 	}
		// 	model.maxLabelLength = max
		// }

	// event dispatched each time a new stream connects to
	// the multiplexer. on-event we need to update the footer
	// model with the new stream information as well as update
	// the Models state. The Model keeps track of connected beams
	// however only cares about the color to use when rendering the logs.
	// Model will ensure that the color for the printed logs of a stream
	// are matching the color information in the footer
	case plexer.Subscriber:

		// // update max label length for indenting
		// // while displaying logs
		// if len(msg) > model.maxLabelLength {
		// 	model.maxLabelLength = len(msg)
		// }

		// label := string(msg)

		// if _, ok := model.beams[label]; !ok {
		// 	color, _ := styles.RandColor()
		// 	model.beams[label] = color

		// 	model.footer, _ = model.footer.Update(subscriber{
		// 		label: label,
		// 		color: color,
		// 	})
		// }

		// model.footer, _ = model.footer.Update(subscriber{
		// 	label: label,
		// 	color: model.beams[label],
		// })

		// return pager, tea.Batch(cmds...)

	// event dispatched by the multiplexer each time a client/stream
	// sends a log linen.
	// The Model needs to add the ansi color code stored for the stream
	// to the dispatched message before adding the data to the ring buffer.
	// Once added to the ring buffer the Model queries for the latest N
	// records (where N is equal to the height of the current viewport.Model)
	// and pass the string to the viewport.Model for rendering
	case plexer.Message:
		color := model.beams[msg.Label]

		space := model.maxLabelLength - len(msg.Label)
		if space < 0 {
			space = 0
		}

		prefix := lipgloss.NewStyle().
			Foreground(color).
			Render(
				msg.Label+strings.Repeat(" ", space),
			) + " | "

		model.buffer.Append(append([]byte(prefix), msg.Data...))

		contents, _ := model.peekBuffer(
			model.height,
			ring.WithLineWrap(model.width),
		)
		model.writer.Reset()
		model.view.SetContent(contents)
		model.view.GotoBottom()
	}

	// propagate event to child models.
	model.view, cmd = model.view.Update(msg)
	cmds = append(cmds, cmd)

	return model, tea.Batch(cmds...)
}

func (model *Model) View() string {
	return lipgloss.JoinVertical(lipgloss.Left,
		model.view.View(),
	)
}

// peekBuffer is a wrapper to read up to N of the last items form the buffer into
// the model.writer. peakBuffer does not reset the model.writer.
func (model Model) peekBuffer(n int, opts ...func(int, []byte) []byte) (string, int) {

	pageSize, err := model.buffer.Peek(
		&model.writer,
		model.height,
		opts...,
	)
	if err != nil {
		debug.Debug(err.Error())
		return "", pageSize
	}

	return model.writer.String(), pageSize
}

func (model *Model) setDimensions(width, height int) {
	model.width, model.height = width-borderMargin, height-bottomSectionHeight-inputSectionHeight-magicNumber
	model.view.Width, model.view.Height = model.width, model.height
}
