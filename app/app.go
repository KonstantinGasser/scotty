package app

import (
	"io"
	"strconv"
	"strings"

	"github.com/KonstantinGasser/scotty/debug"
	"github.com/KonstantinGasser/scotty/ring"

	"github.com/KonstantinGasser/scotty/app/component/formatter"
	"github.com/KonstantinGasser/scotty/app/component/pager"
	"github.com/KonstantinGasser/scotty/app/component/status"
	"github.com/KonstantinGasser/scotty/app/component/welcome"
	"github.com/KonstantinGasser/scotty/app/styles"
	plexer "github.com/KonstantinGasser/scotty/multiplexer"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	cmdUnset = iota
	cmdFormat
	cmdFilter
)

var (
	placeholderDefault = "use \":\" to enter log formatting or \"ctrl+f\" to filter the logs"
	placeholderFormat  = "line number (use k/j to move and ESC/q to exit)"
	placeholderFilter  = "type the name of a stream to highlight its logs"
)

const (
	// scotty has just been started
	// show welcome page
	welcomeView = iota
	// one or more log streams have connected
	// and are streaming logs
	// show logs in tailing window
	tailView
	// a query was execute w/o tailing
	// show view with query results
	formatView

	// space must be reserved for the status either at
	// the top or the bottom however must be taken in
	// account for the other views
	statusHeight = 2

	// positioned at the bottom of the application
	// height of the input model for commands
	inputHeight = 1
)

type App struct {
	quite chan<- struct{}

	keys bindings

	// width and height of the terminal
	// updates as it changes
	width, height int

	// any incoming multiplexer.Message
	// is written to the buffer. However the
	// does not need to read from the buffer, thus
	// only an io.Writer
	buffer io.Writer

	// streams keeps track of all streams which connected
	// to scotty within the same session and store a unique
	// color for each stream. Used when writing and displaying
	// logs from the different streams
	streams map[string]lipgloss.Color

	// tracks the length of the longest stream label currently
	// connected - only used to apply spacing between labels and logs
	maxLabelLength int

	// state can be either of (welcomeView | tailView | formatView)
	// and based on the state View returns the status and the current
	// view from the views map
	state int

	// if true the ":" was hit in the previous message and the user is typing
	awaitInput bool

	// hasInput is true only if awaitInput is true prio and the user has typed
	// in a valid command. If true indicated that a user request is running
	hasInput bool

	// ignoreKey is used to avoid adding certain keys to the input model
	// which would display the keys. Ignored keys are:
	// - "j"
	// - "k"
	ignoreKey bool

	// requestedCommand stores the key which triggered a command and
	// is used to differentiate between follow-up actions after
	// the "enter" key has been hit.
	requestedCommand int

	// views is a map of tea.Model which can be either of type component/pager
	// or component/view and represent the views which should be displayed
	// in the next render
	// TODO: do we want to allow to show multiple views? or only one at the time?
	// background of the question is 1. rending of the string is more complex
	// 2. showing more then one view reduces the space a single view can take.
	// On the other hand in the future it would be cool if the user could store
	// "dashboard" configurations which can be used to pre-initialize scotty
	views map[int]tea.Model

	// display status information. Previously
	// called head or footer however end position
	// is not clear so status is more generic for a name
	status tea.Model

	// input to execute commands
	input textinput.Model

	// errs receives errors happening in the multiplexer
	// while working/reading from streams
	errs <-chan plexer.Error
	// messages receives each message send by a stream,
	// excluding SYNC messages from the client beam command
	messages <-chan plexer.Message
	// beams receives purely information about the fact
	// that a new stream has connected. The received string
	// is the label of the stream
	subscriber  <-chan plexer.Subscriber
	unsubscribe <-chan plexer.Unsubscribe
}

func New(bufferSize int, q chan<- struct{}, errs <-chan plexer.Error, msgs <-chan plexer.Message, subs <-chan plexer.Subscriber, unsubs <-chan plexer.Unsubscribe) (*App, error) {

	// dummy width and height until bubbletea sends first tea.WindowSizeMsg
	width, height := 100, 124

	buffer := ring.New(uint32(bufferSize))

	input := textinput.New()
	input.Placeholder = placeholderDefault
	input.Prompt = ":"

	return &App{
		quite: q,
		keys:  defaultBindings,

		width:  width,
		height: height - statusHeight - inputHeight,

		buffer:  &buffer,
		streams: make(map[string]lipgloss.Color),

		state: welcomeView,
		views: map[int]tea.Model{
			welcomeView: welcome.New(width, height),
			tailView:    pager.New(width, height-statusHeight-inputHeight, &buffer),
			formatView:  formatter.New(width, height-statusHeight-inputHeight, &buffer),
		},
		status: status.New(width, height),

		input:            input,
		awaitInput:       false,
		hasInput:         false,
		ignoreKey:        false,
		requestedCommand: cmdUnset,

		errs:        errs,
		messages:    msgs,
		subscriber:  subs,
		unsubscribe: unsubs,
	}, nil
}

// Init kicks off all the background listening jobs to receive
// tea.Msg coming from outside the app.App such as the multiplexer.Socket
func (app *App) Init() tea.Cmd {
	return tea.Batch(
		app.consumeErrs,
		app.consumeSubscriber,
		app.consumeMsg,
		app.consumerUnsubscribe,
	)
}

func (app *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	var (
		cmds []tea.Cmd
		cmd  tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		// nop, right
		case key.Matches(msg, app.keys.Quit):
			app.quite <- struct{}{}
			return app, tea.Quit
		// indicated that the dev wants to switch the view.
		// Once this key is hit we need to wait for input
		case key.Matches(msg, app.keys.Input):
			app.input.Reset()
			app.input.Placeholder = placeholderFormat
			app.awaitInput = true
			app.input.Focus()
			app.requestedCommand = cmdFormat

			return app, tea.Batch(cmds...)

		case key.Matches(msg, app.keys.Filter):
			app.input.Reset()
			app.input.Placeholder = placeholderFilter
			app.awaitInput = true
			app.input.Focus()
			app.requestedCommand = cmdFilter

			return app, tea.Batch(cmds...)

		// only triggered if input is expected.
		// Evaluates the input and coordinates the required
		// execution. Current allowed input patterns:
		// - [0-9]{1,} -> switch view to formatter with requested input
		// - f:[a-zA-Z_-0-9]{1,} -> add filter on log view
		case key.Matches(msg, app.keys.Enter) && app.awaitInput:

			cmd = app.executeCommand()
			cmds = append(cmds, cmd)

		// propagate event to formatter and request to format
		// previous log line
		case key.Matches(msg, app.keys.Up) && app.state == formatView && app.hasInput:
			cmds = append(cmds, formatter.RequestUp())
			app.ignoreKey = true

		// propagate event to formatter and request to format
		// next log line
		case key.Matches(msg, app.keys.Down) && app.state == formatView && app.hasInput:
			cmds = append(cmds, formatter.RequestDown())
			app.ignoreKey = true

		// terminate formatting view and propagate event to formatter,
		// reinstate tailView as view for app to render
		case key.Matches(msg, app.keys.Exit):
			app.awaitInput = false
			app.input.Blur()
			app.input.Placeholder = placeholderDefault

			cmds = append(cmds, formatter.RequestQuite())
			app.state = tailView
		}

	case tea.WindowSizeMsg:
		app.width = msg.Width
		app.height = msg.Height - statusHeight - inputHeight

		msg.Width = app.width
		msg.Height = app.height

	// event dispatched each time a new stream connects to
	// the multiplexer. on-event we need to update the footer
	// model with the new stream information as well as update
	// the Models state. The Model keeps track of connected beams
	// however only cares about the color to use when rendering the logs.
	// Model will ensure that the color for the printed logs of a stream
	// are matching the color information in the footer
	case plexer.Subscriber:
		// update max label length for indenting
		// while displaying logs
		if len(msg) > app.maxLabelLength {
			app.maxLabelLength = len(msg)
		}

		label := string(msg)

		if _, ok := app.streams[label]; !ok {
			color, _ := styles.RandColor()
			app.streams[label] = color
		}

		app.status, _ = app.status.Update(status.Connection{
			Label: label,
			Color: app.streams[label],
		})
		cmds = append(cmds, app.consumeSubscriber)

		return app, tea.Batch(cmds...)

	// event dispatched each time a beam disconnects from scotty.
	// The message itself is the label of the stream which
	// disconnected. On a disconnect we need to recompute the
	// length of the longest stream label in order to maintain
	// pretty indention for logging the logs with the label prefix
	case plexer.Unsubscribe:
		// we only need to reassign the max value
		// if the current max is disconnecting
		if len(msg) >= app.maxLabelLength {
			max := 0
			for label := range app.streams {
				if len(label) > max && label != string(msg) {
					max = len(label)
				}
			}
			app.maxLabelLength = max
		}
		cmds = append(cmds, app.consumerUnsubscribe)

	// event dispatched by the multiplexer each time a client/stream
	// sends a log linen.
	// The App needs to add the ansi color code stored for the stream
	// to the dispatched message before adding the data to the ring buffer.
	// Further processing happening in active view (views[state]).
	// For references see:
	// - pager.Update()
	case plexer.Message:
		// in all cases the first view we show whenever
		// the first message is received is the pager
		if app.state == welcomeView {
			app.state = tailView
		}

		color := app.streams[msg.Label]

		space := app.maxLabelLength - len(msg.Label)
		if space < 0 {
			space = 0
		}

		prefix := lipgloss.NewStyle().
			Foreground(color).
			Render(
				msg.Label+strings.Repeat(" ", space),
			) + " | "

		app.buffer.Write(append([]byte(prefix), msg.Data...))
		cmds = append(cmds, app.consumeMsg)

	case plexer.Error:
		// good question guess pipe to status???
	}

	// update current model
	app.views[app.state], cmd = app.views[app.state].Update(msg)
	cmds = append(cmds, cmd)

	if !app.ignoreKey {
		app.input, cmd = app.input.Update(msg)
		cmds = append(cmds, cmd)
	}
	app.ignoreKey = false

	return app, tea.Batch(cmds...)
}

func (app *App) View() string {
	if app.state == welcomeView {
		return app.views[app.state].View()
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		app.status.View(),
		app.views[app.state].View(),
		app.input.View(),
	)
}

// executeCommands interprets the users input
// and based on the previous keystroke a follow-up
// action in form of an tea.Cmd is evaluated.
func (app *App) executeCommand() tea.Cmd {

	value := app.input.Value()

	switch app.requestedCommand {
	case cmdFormat:
		index, err := strconv.Atoi(value)
		if err != nil {
			debug.Print("unable to parse index: %q: %w", value, err)
		}
		app.hasInput = true

		app.state = formatView
		return formatter.RequestView(index)

	case cmdFilter:
		debug.Print("filter: %s\n", value)
	}
	return nil
}

/* consume* yields back a tea.Msg piped through a channel ending in the app.Update func */
func (app *App) consumeMsg() tea.Msg          { return <-app.messages }
func (app *App) consumeErrs() tea.Msg         { return <-app.errs }
func (app *App) consumeSubscriber() tea.Msg   { return <-app.subscriber }
func (app *App) consumerUnsubscribe() tea.Msg { return <-app.unsubscribe }
