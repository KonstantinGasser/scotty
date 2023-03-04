package formatter

import (
	"bytes"

	"github.com/KonstantinGasser/scotty/app/styles"
	"github.com/KonstantinGasser/scotty/debug"
	"github.com/KonstantinGasser/scotty/ring"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	borderMargin = 1

	awaitInput = iota + 1
	hasInput
)

var (
	modelStyle = lipgloss.NewStyle().Padding(0, 1)

	inputStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(styles.DefaultColor.Border)
)

type subscriber struct {
	label string
	color lipgloss.Color
}

type requestedView int
type requestedUp struct{}
type requestedDown struct{}
type requestedQuite struct{}

func RequestView(index int) tea.Cmd {
	return func() tea.Msg {
		return requestedView(index)
	}
}

func RequestUp() tea.Cmd {
	return func() tea.Msg {
		return requestedUp{}
	}
}

func RequestDown() tea.Cmd {
	return func() tea.Msg {
		return requestedDown{}
	}
}

func RequestQuite() tea.Cmd {
	return func() tea.Msg {
		return requestedQuite{}
	}
}

// Model implements the tea.Model interface.
// Furthermore, Model allows to tail logs.
// Model does not not store the logs its only
// porose is it to display them.
type Model struct {
	ready bool

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

	// relativeIndex represents the requested index
	// on the current page. Whichever item == index
	// is highlighted and formatted. Note relativeIndex
	// is the relative index to the page. In order
	// to get the index of the element within the
	// buffer one must add the offsetStart to the
	// relativeIndex to get the absolute index.
	relativeIndex int

	// absoluteIndex refers to the actual index in the
	// buffer which is currently formatted
	absoluteIndex int

	// offsetStart if used when paging through the logs
	// and formatting log lines. It refers to the index
	// with which the model starts (first log of the page)
	offsetStart int

	// pageSize refers to the number of items currently
	// visible in the view - line wraps are not included
	// an item which takes up two lines counts as one
	pageSize int
}

func New(buffer *ring.Buffer) *Model {

	return &Model{
		ready: false,

		buffer: buffer,
		writer: bytes.Buffer{},

		beams:          map[string]lipgloss.Color{},
		maxLabelLength: 0,
		view:           viewport.Model{},
		width:          0,
		height:         0,
		relativeIndex:  -1,
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

	case requestedView:
		model.initFormattingMode(int(msg))

		model.reloadPage(
			model.absoluteIndex,
			ring.WithInlineFormatting(model.width, model.absoluteIndex),
			ring.WithLineWrap(model.width),
		)
		model.view.GotoTop()

	case requestedQuite:
		model.resetFormattingMode()

		model.writer.Reset()
		model.view.SetContent("") // unset content
		return model, tea.Batch(cmds...)

	// selects the previous log line to be parsed
	// and displayed. Input ignores when relativeIndex <= 0
	case requestedUp:
		if model.absoluteIndex == 0 {
			break
		}

		model.relativeIndex--
		model.absoluteIndex--

		// requested index to format is outside (above)
		// the current view as such we need to shift the
		// content of the view up. For scrolling up we
		// only show the next upper element not an entire
		// new page like we do when scrolling down.
		if model.relativeIndex < 0 {

			model.moveUp(
				ring.WithInlineFormatting(model.width, model.absoluteIndex),
				ring.WithLineWrap(model.width),
			)

			break
		}

		model.reloadPage(
			model.absoluteIndex,
			ring.WithInlineFormatting(model.width, model.absoluteIndex),
			ring.WithLineWrap(model.width),
		)

		// for this key stroke we don't need the msg any other where
		// and we putted to the input model the stork is registered
		// which we don't want
		return model, tea.Batch(cmds...)

	// selects the next log line to be parsed and
	// displayed. Input ignored when relativeIndex >= buffer.cap
	case requestedDown:
		model.relativeIndex++ // index of the within the current page
		model.absoluteIndex++ // overall index of the selected item in the buffer

		// nil items in the buffer indicated that the buffer is not full
		// and the requested index exists but has not been written to yet.
		// Just means user wanted a log that has not been beamed yet.
		if model.buffer.Nil(model.absoluteIndex) {
			// well showing nothing is not cool
			// compensate to last working index
			model.absoluteIndex--
			model.relativeIndex--

			model.reloadPage(
				model.absoluteIndex,
				ring.WithInlineFormatting(model.width, model.absoluteIndex),
				ring.WithLineWrap(model.width),
			)
			break
		}

		// check if the requested log line is out of
		// the view (not included in the previous render)
		// if so we need to adjust the page/go to the next
		// page an rerender the view again
		if model.relativeIndex >= model.pageSize {

			model.moveDown(
				ring.WithInlineFormatting(model.width, model.absoluteIndex),
				ring.WithLineWrap(model.width),
			)

			break
		}

		// render logs starting from the offset
		// till offset+height. The returned
		// lines indicate how many items are included
		// in the render (hard to tell solely based on the string
		// from the model.writer)
		model.reloadPage(
			model.absoluteIndex,
			ring.WithInlineFormatting(model.width, model.absoluteIndex),
			ring.WithLineWrap(model.width),
		)

		// for this key stroke we don't need the msg any other where
		// and we putted to the input model the stork is registered
		// which we don't want
		return model, tea.Batch(cmds...)

	// event dispatched from bubbletea when the screen size changes.
	// We need to update the model and model.view width and height.
	// However, if the parsing mode is on the width is only 2/3
	// of the available screen size.
	case tea.WindowSizeMsg:

		if !model.ready {
			model.width = msg.Width - borderMargin
			model.height = styles.AvailableHeight(msg.Height)

			model.view = viewport.New(model.width-borderMargin, model.height)
			model.view.Width = model.width - borderMargin
			model.view.Height = model.height
			model.view.MouseWheelEnabled = true

			model.ready = true
			break
		}

		model.setDimensions(
			msg.Width,
			msg.Height,
		)

		model.reloadPage(
			model.absoluteIndex,
			ring.WithInlineFormatting(model.width, model.absoluteIndex),
			ring.WithLineWrap(model.width),
		)
	}

	// propagate event to child models.
	model.view, cmd = model.view.Update(msg)
	cmds = append(cmds, cmd)

	return model, tea.Batch(cmds...)
}

func (model *Model) View() string {

	if !model.ready {
		return "initializing..."
	}

	return model.view.View()
}

// moveUp renders the current window shifted up by 1
// and selects the relativeIndex of zero.
func (model *Model) moveUp(opts ...func(int, []byte) []byte) {

	if model.offsetStart == 0 {
		return
	}

	model.relativeIndex = 0
	model.offsetStart--
	if model.offsetStart < 0 {
		model.offsetStart = 0
	}

	contents, pageSize := model.offsetBuffer(
		model.offsetStart,
		model.height,
		ring.WithInlineFormatting(model.width, model.absoluteIndex),
		ring.WithLineWrap(model.width),
	)
	model.writer.Reset()
	model.pageSize = pageSize
	model.view.SetContent(contents)
}

// reloadPage requests the same offset window from the buffer
// however with a different line marked as selected
func (model *Model) reloadPage(selected int, opts ...func(int, []byte) []byte) {

	contents, pageSize := model.offsetBuffer(
		model.offsetStart,
		model.height,
		opts...,
	)
	model.writer.Reset()
	model.pageSize = pageSize

	model.view.SetContent(contents)
}

// moveDown "scrolls" the buffer down by one page defined bei the current viewport height.
// The relativeIndex referring to the line in the current window selected is set to zero
// again.
func (model *Model) moveDown(opts ...func(int, []byte) []byte) {

	model.offsetStart += model.relativeIndex
	// reset relativeIndex since its relative to
	// the current page. When the page changes
	// the relative index is 0
	model.relativeIndex = 0

	contents, pageSize := model.offsetBuffer(
		model.offsetStart,
		model.height,
		opts...,
	)
	model.writer.Reset()

	model.pageSize = pageSize

	model.view.SetContent(contents)
}

func (model Model) offsetBuffer(start, end int, opts ...func(int, []byte) []byte) (string, int) {

	pageSize, err := model.buffer.ReadOffset(
		&model.writer,
		start,
		end,
		opts...,
	)
	if err != nil {
		debug.Debug(err.Error())
		return "", pageSize
	}

	return model.writer.String(), pageSize
}

func (model *Model) setDimensions(width, height int) {
	model.width, model.height = width-borderMargin, styles.AvailableHeight(height)
	model.view.Width, model.view.Height = model.width, model.height
}

func (model *Model) initFormattingMode(offset int) {
	model.offsetStart = offset
	model.absoluteIndex = model.offsetStart
	model.relativeIndex = 0
}

func (model *Model) resetFormattingMode() {
	model.offsetStart = -1
	model.relativeIndex = -1
	model.absoluteIndex = -1
}
