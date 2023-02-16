package app

import (
	"fmt"

	"github.com/KonstantinGasser/scotty/ring"

	"github.com/KonstantinGasser/scotty/app/component/formatter"
	"github.com/KonstantinGasser/scotty/app/component/pager"
	"github.com/KonstantinGasser/scotty/app/component/welcome"
	plexer "github.com/KonstantinGasser/scotty/multiplexer"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
)

type App struct {
	quite chan<- struct{}

	keys bindings

	// width and height of the terminal
	// updates as it changes
	width, height int

	// streams keeps track of all streams which connected
	// to scotty within the same session and store a unique
	// color for each stream. Used when writing and displaying
	// logs from the different streams
	streams map[string]lipgloss.Color

	// state can be either of (welcomeView | tailView | formatView)
	// and based on the state View returns the status and the current
	// view from the views map
	state int

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

	width, height, err := windowSize()
	if err != nil {
		return nil, fmt.Errorf("unable to determine the initial dimensions of the terminal: %w", err)
	}

	buffer := ring.New(uint32(bufferSize))

	return &App{
		quite: q,
		keys:  defaultBindings,

		width: width,

		streams: make(map[string]lipgloss.Color),

		state: welcomeView,
		views: map[int]tea.Model{
			welcomeView: welcome.New(width, height),
			tailView:    pager.New(width, height, &buffer),
			formatView:  formatter.New(width, height, &buffer),
		},

		height: height,

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
		if quite := app.resolveKey(msg); quite != nil {
			app.quite <- struct{}{}
			return app, quite
		}
	case tea.WindowSizeMsg:
		app.width = msg.Width
		app.height = msg.Height

	// currently the app is not doing anything with the message as its child models are taking care
	// of it. However, the app requests bubbletea to wait and listen for new messages pushed to
	// channels - this ends now!
	case plexer.Message, plexer.Error, plexer.Subscriber, plexer.Unsubscribe:

		switch msg.(type) {
		case plexer.Error:
			cmds = append(cmds, app.consumeErrs)
		case plexer.Subscriber:
			cmds = append(cmds, app.consumeSubscriber)
		case plexer.Unsubscribe:
			cmds = append(cmds, app.consumerUnsubscribe)
		case plexer.Message:
			cmds = append(cmds, app.consumeMsg)
		}

		// enable tailing of logs view
		if app.state == welcomeView {
			app.state = tailView
		}
	}

	// update current model
	app.views[app.state], cmd = app.views[app.state].Update(msg)
	cmds = append(cmds, cmd)

	return app, tea.Batch(cmds...)
}

func (app *App) View() string {
	return app.views[app.state].View()
}

/* consume* yields back a tea.Msg piped through a channel ending in the app.Update func */
func (app *App) consumeMsg() tea.Msg          { return <-app.messages }
func (app *App) consumeErrs() tea.Msg         { return <-app.errs }
func (app *App) consumeSubscriber() tea.Msg   { return <-app.subscriber }
func (app *App) consumerUnsubscribe() tea.Msg { return <-app.unsubscribe }
