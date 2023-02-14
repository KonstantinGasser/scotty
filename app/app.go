package app

import (
	"fmt"

	"github.com/KonstantinGasser/scotty/app/component/pager"
	"github.com/KonstantinGasser/scotty/app/component/welcome"
	plexer "github.com/KonstantinGasser/scotty/multiplexer"
	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	// scotty has just been started
	// show welcome page
	welcomeView = iota
	// one or more log streams have connected
	// and are streaming logs
	// show logs in tailing window
	logTailView
	// a query was execute w/o tailing
	// show view with query results
	queryView
	// a query was executed with tailing
	// show logs in tailing window with query filters
	queryTailView
)

type App struct {
	quite chan<- struct{}

	// help view
	help help.Model

	keys bindings

	// views is a map of tea.Model which can be either of type component/pager
	// or component/view and represent the views which should be displayed
	// in the next render
	// TODO: do we want to allow to show multiple views? or only one at the time?
	// background of the question is 1. rending of the string is more complex
	// 2. showing more then one view reduces the space a single view can take.
	// On the other hand in the future it would be cool if the user could store
	// "dashboard" configurations which can be used to pre-initialize scotty
	views map[int]tea.Model

	// width and height of the terminal
	// updates as it changes
	width, height int
	state         int

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

	return &App{
		quite: q,
		help:  help.New(),
		keys:  defaultBindings,

		views: map[int]tea.Model{
			welcomeView: welcome.New(width, height),
			logTailView: pager.New(width, height, bufferSize), // have this pre-initialized as it will be need no matter what
		},

		width:  width,
		height: height,
		state:  welcomeView,

		errs:        errs,
		messages:    msgs,
		subscriber:  subs,
		unsubscribe: unsubs,
	}, nil
}

/* consume* yields back a tea.Msg piped through a channel ending in the app.Update func */
func (app *App) consumeMsg() tea.Msg          { return <-app.messages }
func (app *App) consumeErrs() tea.Msg         { return <-app.errs }
func (app *App) consumeSubscriber() tea.Msg   { return <-app.subscriber }
func (app *App) consumerUnsubscribe() tea.Msg { return <-app.unsubscribe }

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
	// channels
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
			app.state = logTailView
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
