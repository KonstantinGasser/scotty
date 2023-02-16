package formatter

import (
	"bytes"
	"strconv"

	"github.com/KonstantinGasser/scotty/app/styles"
	"github.com/KonstantinGasser/scotty/debug"
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
	modelStyle = lipgloss.NewStyle().Padding(0, 1)

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

	// awaitInput indicated if ECS is pressed.
	// if awaitInput == false the input for commands
	// is focused else moved out of focus
	awaitInput bool

	// input is the input field to select
	// an index to format and input further
	// commands
	input textinput.Model

	// some characters inputted we don't want to
	// propagate down to the textinput.Model
	// as they are treated as regular chars
	// and displayed as value - as such indicated if
	// propagation should be ignored
	ignoreInput bool
}

func New(width, height int, buffer *ring.Buffer) *Model {
	w, h := width-borderMargin, height-bottomSectionHeight-inputSectionHeight-magicNumber

	view := viewport.New(w, h)
	view.Height = h
	view.Width = w
	view.MouseWheelEnabled = true
	view.Style = modelStyle.Width(w)

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
		awaitInput:     false,
		relativeIndex:  -1,
		input:          input,
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
		switch msg.String() {
		// triggers the parsing mode of logs. Has no
		// effect while in parsing mode (awaitInput == true)
		case ":":
			model.awaitInput = true

			// the char ":" is not a cmd for the textinput.Model
			// and if passed to the update func of the model is
			// added to the input - which is want we don't want
			model.ignoreInput = true
			model.input.Focus()

		case "enter":
			if !model.awaitInput {
				break
			}

			value := model.input.Value()
			index, err := strconv.Atoi(value)
			if err != nil {
				debug.Print("input %q is not numeric. Type the index of the line you want to parse", value)
				break
			}

			model.initFormattingMode(index)

			model.updatePage(
				model.absoluteIndex,
				ring.WithInlineFormatting(model.width, model.absoluteIndex),
				ring.WithLineWrap(model.width),
			)
			model.view.GotoTop()

			model.input.Blur()
			model.input.Reset()

		// exits the parsing mode. Has no effect
		// while not in parsing mode (awaitInput == false)
		case "esc", "q":
			if !model.awaitInput {
				break
			}

			model.resetFormattingMode()

			// width, height, err := styles.WindowSize()
			// if err != nil {
			// 	debug.Print("[tea.KeyMsg(esc)] unable to get tty width and height: %w\n", err)
			// }

			// model.setDimensions(
			// 	width,
			// 	height,
			// )

			// again the width of the log view changes on
			// exit as such we need to force a rerender
			// in order to fix the line wraps of each log
			// contents, _ := model.peekBuffer(
			// 	model.height,
			// 	ring.WithLineWrap(model.width),
			// )
			model.writer.Reset()
			model.view.SetContent("") // unset content
			// model.view.SetContent(contents)
			// model.view.GotoBottom()

		// selects the previous log line to be parsed
		// and displayed. Input ignores when relativeIndex <= 0
		case "k":

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

				model.previousPage(
					ring.WithInlineFormatting(model.width, model.absoluteIndex),
					ring.WithLineWrap(model.width),
				)

				break
			}

			model.updatePage(
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
		case "j":
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

				model.updatePage(
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

				model.nextPage(
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
			model.updatePage(
				model.absoluteIndex,
				ring.WithInlineFormatting(model.width, model.absoluteIndex),
				ring.WithLineWrap(model.width),
			)

			// for this key stroke we don't need the msg any other where
			// and we putted to the input model the stork is registered
			// which we don't want
			return model, tea.Batch(cmds...)
		}

	// event dispatched from bubbletea when the screen size changes.
	// We need to update the model and model.view width and height.
	// However, if the parsing mode is on the width is only 2/3
	// of the available screen size.
	case tea.WindowSizeMsg:

		model.setDimensions(
			msg.Width,
			msg.Height,
		)

		if model.awaitInput && model.relativeIndex >= 0 {

			model.updatePage(
				model.absoluteIndex,
				ring.WithInlineFormatting(model.width, model.absoluteIndex),
				ring.WithLineWrap(model.width),
			)

			break
		}

		model.updatePage(
			model.absoluteIndex,
			ring.WithInlineFormatting(model.width, model.absoluteIndex),
			ring.WithLineWrap(model.width),
		)
		model.view.GotoBottom()

	}

	// propagate event to child models.
	model.view, cmd = model.view.Update(msg)
	cmds = append(cmds, cmd)

	// only there to avoid certain chars to be used as
	// input for the input field.
	// chars include: "j", "k"
	if !model.ignoreInput {
		model.input, cmd = model.input.Update(msg)
		cmds = append(cmds, cmd)
	}
	// ignoring input is only valid for one pass
	// revert it before the next pass
	model.ignoreInput = false

	return model, tea.Batch(cmds...)
}

func (model *Model) View() string {
	return lipgloss.JoinVertical(lipgloss.Left,
		model.view.View(),
		model.input.View(),
	)
}

// previousPage renders the current window shifted up by 1
// and selects the relativeIndex of zero.
func (model *Model) previousPage(opts ...func(int, []byte) []byte) {

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

// updatePage requests the same offset window from the buffer
// however with a different line marked as selected
func (model *Model) updatePage(selected int, opts ...func(int, []byte) []byte) {

	contents, pageSize := model.offsetBuffer(
		model.offsetStart,
		model.height,
		opts...,
	)
	model.writer.Reset()
	model.pageSize = pageSize

	model.view.SetContent(contents)
}

// nextPage "scrolls" the buffer down by one page defined bei the current viewport height.
// The relativeIndex referring to the line in the current window selected is set to zero
// again.
func (model *Model) nextPage(opts ...func(int, []byte) []byte) {

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

	pageSize, err := model.buffer.Offset(
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
	model.width, model.height = width-borderMargin, height-bottomSectionHeight-inputSectionHeight-magicNumber
	model.view.Width, model.view.Height = model.width, model.height
}

func (model *Model) initFormattingMode(offset int) {
	model.offsetStart = offset
	model.absoluteIndex = model.offsetStart
	model.relativeIndex = 0
}

func (model *Model) resetFormattingMode() {
	model.awaitInput = false
	model.offsetStart = -1
	model.relativeIndex = -1
	model.absoluteIndex = -1

	model.input.Reset()
	model.input.Blur()
}
