package pager

import (
	"bytes"

	"github.com/KonstantinGasser/scotty/app/event"
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
	borderMargin = 1

	// wow literally no idea why this number hence
	// the variable name - if you get why tell me and
	// pls open a PR..else pls don't change it
	magicNumber = 2
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

	// offsetStart if used when paging through the logs
	// and formatting log lines. It refers to the index
	// with which the pager starts (first log of the page)
	offsetStart int

	// pageSize refers to the number of items currently
	// visible in the view - line wraps are not included
	// an item which takes up two lines counts as one
	pageSize int
}

func New(buffer *ring.Buffer) *Model {

	input := textinput.New()
	input.Placeholder = "line number (use k/j to move and ESC/q to exit)"
	input.Prompt = ":"

	return &Model{
		ready:  false,
		buffer: buffer,
		writer: bytes.Buffer{},

		beams:          map[string]lipgloss.Color{},
		maxLabelLength: 0,
		view:           viewport.Model{},
		width:          0,
		height:         0,
	}
}

func (model *Model) Init() tea.Cmd {
	return nil
}

func (model *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	var (
		cmds []tea.Cmd
		cmd  tea.Cmd
		err  error
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
		if !model.ready {
			model.width = msg.Width - borderMargin
			model.height = styles.AvailableHeight(msg.Height)

			model.view = viewport.New(model.width, model.height)
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

		model.pageSize, err = model.buffer.Read(
			&model.writer,
			model.height,
			ring.WithLineWrap(model.width),
		)
		if err != nil {
			debug.Debug(err.Error())
		}

		model.view.SetContent(model.writer.String())
		model.writer.Reset()
		model.view.GotoBottom()

	case event.ReloadBuffer:
		model.pageSize, err = model.buffer.Read(
			&model.writer,
			model.height,
			ring.WithLineWrap(model.width),
		)
		if err != nil {
			debug.Print("[pager.Update(event.ReloadBuffer)] error: %v\n", err)
		}

		model.view.SetContent(model.writer.String())
		model.writer.Reset()
		model.view.GotoBottom()

	// event dispatched by the multiplexer each time a client/stream
	// sends a log linen.
	// The Model needs to add the ansi color code stored for the stream
	// to the dispatched message before adding the data to the ring buffer.
	// Once added to the ring buffer the Model queries for the latest N
	// records (where N is equal to the height of the current viewport.Model)
	// and pass the string to the viewport.Model for rendering
	case plexer.Message:

		model.pageSize, err = model.buffer.Read(
			&model.writer,
			model.height,
			ring.WithLineWrap(model.width),
		)
		if err != nil {
			debug.Print("[pager.Update(plexer.Message)] error: %v\n", err)
		}

		model.view.SetContent(model.writer.String())
		model.writer.Reset()
		model.view.GotoBottom()
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

func (model *Model) setDimensions(width, height int) {
	model.width, model.height = width-borderMargin, styles.AvailableHeight(height)
	model.view.Width, model.view.Height = model.width, model.height
}
