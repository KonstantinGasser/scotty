package app

import (
	"bytes"
	"strconv"
	"strings"

	"github.com/KonstantinGasser/scotty/app/component/browsing"
	"github.com/KonstantinGasser/scotty/app/component/info"
	"github.com/KonstantinGasser/scotty/app/component/querying"
	"github.com/KonstantinGasser/scotty/app/component/tailing"
	"github.com/KonstantinGasser/scotty/app/component/welcome"
	"github.com/KonstantinGasser/scotty/app/styles"
	"github.com/KonstantinGasser/scotty/debug"
	"github.com/KonstantinGasser/scotty/multiplexer"
	"github.com/KonstantinGasser/scotty/store"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	tabItems       = []string{"(1) follow logs", "(2) browse logs", "(3) query logs", "(4) docs"}
	defaultTabLine = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderTop(false).BorderLeft(false).BorderRight(false).
			BorderBottom(true).
			Render(
			lipgloss.JoinHorizontal(lipgloss.Left,
				styles.ActiveTab(tabItems[tabFollow]),
				styles.Tab(tabItems[tabBrowse]),
				styles.Tab(tabItems[tabQuery]),
				styles.Tab(tabItems[tabDocs]),
			),
		)
)

const (
	tabUnset = iota - 1
	tabFollow
	tabBrowse
	tabQuery
	tabDocs
)

type streamConfig struct {
	color lipgloss.Color
}

type App struct {
	/* internal properties */
	// indication to close and stop work.
	// Signal send from outside
	quite chan<- struct{}
	// availabel dimension
	ttyWidth, ttyHeight int
	// false until initial tea.WindowSizeMsg send
	// and App is initialized
	ready bool
	// key bindings
	bindings bindings

	/* multiplexer / i/o properties */
	// channels to consume multiplexer events
	consumer      multiplexer.Consumer
	streamConfigs map[string]streamConfig

	// place where all logs are written
	// to. App manly uses it for inserts
	logstore *store.Store

	/* layout properties */
	// finished parsed and build tabs
	// where one tab is shown as active.
	// Default is welcome-tab:active
	tabLine string
	// indication which tab is currently
	// active and thereby which component
	activeTab int

	/* component specific properties */
	infoComponent tea.Model
	// map of all available components mapped to
	// the available tabs
	compontens map[int]tea.Model
}

func New(q chan<- struct{}, lStore *store.Store, consumer multiplexer.Consumer) *App {

	return &App{
		quite:     q,
		ttyWidth:  -1, // unset/invalid
		ttyHeight: -1, // unset/invalid
		ready:     false,
		bindings:  defaultBindings,

		consumer:      consumer,
		streamConfigs: make(map[string]streamConfig),
		logstore:      lStore,

		tabLine:   defaultTabLine,
		activeTab: tabUnset,

		infoComponent: info.New(),
		compontens: map[int]tea.Model{
			tabFollow: tailing.New(lStore.NewPager(0, 0)),
			tabBrowse: browsing.New(),
			tabQuery:  querying.New(),
		},
	}
}

func (app App) Init() tea.Cmd {
	return tea.Batch(
		app.consumeMsg,
		app.consumeSubscriber,
		app.consumerUnsubscribe,
		app.consumeErrs,
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
		case key.Matches(msg, app.bindings.Quit):
			app.quite <- struct{}{}
			return app, tea.Quit
		case key.Matches(msg, app.bindings.SwitchTab):
			tabIndex, _ := strconv.ParseInt(msg.String(), 10, 64)
			tabIndex = tabIndex - 1 // -1 as it is displayed as 1 2 3 4 but index at 0
			debug.Print("[app] switch to tab (%d)\n", tabIndex)

			if app.activeTab == int(tabIndex) {
				return app, tea.Batch(cmds...)
			}

			app.activeTab = int(tabIndex)
			app.updateActiveTab()

		}
		// update active componten to handle key store individually
		app.compontens[app.activeTab], cmd = app.compontens[app.activeTab].Update(msg)
		return app, tea.Batch(cmds...)
	case tea.WindowSizeMsg:
		app.ttyWidth, app.ttyHeight = msg.Width, msg.Height

		app.infoComponent, cmd = app.infoComponent.Update(msg)
		cmds = append(cmds, cmd)

		// iterate over all compontens as they are not
		// aware of the inital width and height of the tty
		if !app.ready {

			for i, comp := range app.compontens {
				app.compontens[i], cmd = comp.Update(msg)
				cmds = append(cmds, cmd)
			}

			app.ready = true
			return app, tea.Batch(cmds...)
		}

		app.compontens[app.activeTab], cmd = app.compontens[app.activeTab].Update(msg)
		cmds = append(cmds, cmd)

	// triggered each time a new stream connects successfully to scotty and is procssed
	// by the multiplexer. A random color is assigned to the stream if not yet pressent
	// (identified by its label). An update about the new stream is propagated to the info
	// component.
	case multiplexer.Subscriber:
		if _, ok := app.streamConfigs[string(msg)]; ok {
			break
		}

		fg, _ := styles.RandColor()
		app.streamConfigs[string(msg)] = streamConfig{color: fg}
		app.infoComponent.Update(info.NewBeam(string(msg), fg))

		cmds = append(cmds, app.consumeSubscriber)

	// triggered each time a new message is pushed from the multiplexer to
	// the consumer.
	// Requires to identify the stream the message is from, build the prefix
	// and to store the message in the log-store. Furthermore, inserts into
	// the log-store will happend dispite the active tab. This allows background
	// updates of the follow-components between tab switches.
	case multiplexer.Message:
		if app.activeTab == tabUnset {
			app.activeTab = tabFollow
			app.updateActiveTab()
		}

		config, ok := app.streamConfigs[msg.Label]
		if !ok {
			break
		}

		prefix := lipgloss.NewStyle().Foreground(config.color).Render(msg.Label) + " | "

		app.logstore.Insert(msg.Label, len(prefix), append([]byte(prefix), bytes.TrimSpace(msg.Data)...))
		// update follow component asap in order to allow background updates while
		// in a different tab
		app.compontens[tabFollow], _ = app.compontens[tabFollow].Update(msg)
		cmds = append(cmds, app.consumeMsg)
	}

	// follow component is updates asap after a message is received
	if app.activeTab != tabUnset && app.activeTab != tabFollow {
		app.compontens[app.activeTab], cmd = app.compontens[app.activeTab].Update(msg)
		cmds = append(cmds, cmd)
	}

	return app, tea.Batch(cmds...)
}

func (app App) View() string {

	if app.activeTab == tabUnset {
		return welcome.New(app.ttyWidth, app.ttyHeight).View()
	}

	return lipgloss.NewStyle().
		Padding(styles.ContentPaddingVertical, 0).
		Render(
			lipgloss.JoinVertical(lipgloss.Left,
				app.tabLine,
				app.compontens[app.activeTab].View(),
				app.infoComponent.View(),
			),
		)
}

func (app *App) updateActiveTab() {
	items := make([]string, len(tabItems))
	for i, label := range tabItems {
		if i == app.activeTab {
			items[i] = styles.ActiveTab(label)
			continue
		}
		items[i] = styles.Tab(label)
	}
	tabs := lipgloss.JoinHorizontal(lipgloss.Left, items...)
	app.tabLine = lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderTop(false).BorderLeft(false).BorderRight(false).
		BorderBottom(true).
		Render(
			tabs + strings.Repeat(" ", app.ttyWidth-lipgloss.Width(tabs)),
		)
}

/* consume* yields back a tea.Msg piped through a channel ending in the app.Update func */
func (app *App) consumeMsg() tea.Msg          { return <-app.consumer.Messages() }
func (app *App) consumeErrs() tea.Msg         { return <-app.consumer.Errors() }
func (app *App) consumeSubscriber() tea.Msg   { return <-app.consumer.Subscribers() }
func (app *App) consumerUnsubscribe() tea.Msg { return <-app.consumer.Unsubscribers() }

// const (
// 	cmdUnset = iota
// 	cmdFormat
// 	cmdFilter
// )

// var (
// 	placeholderDefault = "use \":\" to enter log formatting or \"ctrl+f\" to filter the logs"
// 	placeholderFormat  = "line number (use k/j to move and ESC/q to exit)"
// 	placeholderFilter  = "type the name of a stream to highlight its logs"

// 	promptDefault = ""
// 	promptFormat  = lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Render("index: ")
// 	promptFilter  = lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Render("filter: ")
// )

// const (
// 	// first tea.WindowSizeMsg has not been send yet
// 	initializing = iota
// 	// scotty has just been started
// 	// show welcome page
// 	welcomeView
// 	// one or more log streams have connected
// 	// and are streaming logs
// 	// show logs in tailing window
// 	tailView
// 	// a query was execute w/o tailing
// 	// show view with query results
// 	formatView
// )

// type App struct {
// 	quite chan<- struct{}

// 	keys bindings

// 	// width and height of the terminal
// 	// updates as it changes
// 	width, height int

// 	// any incoming multiplexer.Message
// 	// is written to the buffer. However the
// 	// does not need to read from the buffer, thus
// 	// only an io.Writer
// 	buffer   *ring.Buffer
// 	logstore *store.Store

// 	// streams keeps track of all streams which connected
// 	// to scotty within the same session and store a unique
// 	// color for each stream. Used when writing and displaying
// 	// logs from the different streams
// 	streams map[string]lipgloss.Color

// 	// tracks the length of the longest stream label currently
// 	// connected - only used to apply spacing between labels and logs
// 	maxLabelLength int

// 	// state can be either of (welcomeView | tailView | formatView)
// 	// and based on the state View returns the status and the current
// 	// view from the views map
// 	state int

// 	// if true the ":" was hit in the previous message and the user is typing
// 	awaitInput bool

// 	// hasInput is true only if awaitInput is true prio and the user has typed
// 	// in a valid command. If true indicated that a user request is running
// 	hasInput bool

// 	// ignoreKey is used to avoid adding certain keys to the input model
// 	// which would display the keys. Ignored keys are:
// 	// - "j"
// 	// - "k"
// 	ignoreKey bool

// 	// requestedCommand stores the key which triggered a command and
// 	// is used to differentiate between follow-up actions after
// 	// the "enter" key has been hit.
// 	requestedCommand int

// 	// views is a map of tea.Model which can be either of type component/pager
// 	// or component/view and represent the views which should be displayed
// 	// in the next render
// 	// TODO: do we want to allow to show multiple views? or only one at the time?
// 	// background of the question is 1. rending of the string is more complex
// 	// 2. showing more then one view reduces the space a single view can take.
// 	// On the other hand in the future it would be cool if the user could store
// 	// "dashboard" configurations which can be used to pre-initialize scotty
// 	views map[int]tea.Model

// 	// display status information. Previously
// 	// called head or footer however end position
// 	// is not clear so status is more generic for a name
// 	status tea.Model

// 	// input to execute commands
// 	input textinput.Model

// 	// errs receives errors happening in the multiplexer
// 	// while working/reading from streams
// 	errs <-chan plexer.Error
// 	// messages receives each message send by a stream,
// 	// excluding SYNC messages from the client beam command
// 	messages <-chan plexer.Message
// 	// beams receives purely information about the fact
// 	// that a new stream has connected. The received string
// 	// is the label of the stream
// 	subscriber  <-chan plexer.Subscriber
// 	unsubscribe <-chan plexer.Unsubscribe
// }

// func New(bufferSize int, q chan<- struct{}, errs <-chan plexer.Error, msgs <-chan plexer.Message, subs <-chan plexer.Subscriber, unsubs <-chan plexer.Unsubscribe) (*App, error) {

// 	buffer := ring.New(uint32(bufferSize))

// 	logstore := store.New(uint32(bufferSize))

// 	input := textinput.New()
// 	input.Placeholder = placeholderDefault
// 	input.Prompt = promptDefault

// 	return &App{
// 		quite: q,
// 		keys:  defaultBindings,

// 		width:  0,
// 		height: 0,

// 		buffer:   &buffer,
// 		logstore: logstore,
// 		streams:  make(map[string]lipgloss.Color),

// 		state: initializing,
// 		views: map[int]tea.Model{
// 			welcomeView: welcome.New(),
// 			tailView:    tailing.New(logstore.NewPager(55, 100)), //pager.New(&buffer),
// 			formatView:  formatter.New(&buffer),
// 		},
// 		status: status.New(),

// 		input:            input,
// 		awaitInput:       false,
// 		hasInput:         false,
// 		ignoreKey:        false,
// 		requestedCommand: cmdUnset,

// 		errs:        errs,
// 		messages:    msgs,
// 		subscriber:  subs,
// 		unsubscribe: unsubs,
// 	}, nil
// }

// // Init kicks off all the background listening jobs to receive
// // tea.Msg coming from outside the app.App such as the multiplexer.Socket
// func (app *App) Init() tea.Cmd {
// 	return tea.Batch(
// 		app.consumeErrs,
// 		app.consumeSubscriber,
// 		app.consumeMsg,
// 		app.consumerUnsubscribe,
// 	)
// }

// func (app *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

// 	var (
// 		cmds []tea.Cmd
// 		cmd  tea.Cmd
// 	)

// 	if app.state == initializing {

// 		window, ok := msg.(tea.WindowSizeMsg)
// 		if !ok {
// 			return app, nil
// 		}

// 		app.width = window.Width
// 		app.height = styles.AvailableHeight(window.Height)

// 		// before we can start we need to wait for bubbletea to
// 		// send the first tea.WindowSizeMsg and we must propagate
// 		// the message to all views + status in order to initialize them
// 		if app.state == initializing {
// 			for view, model := range app.views {
// 				app.views[view], cmd = model.Update(msg)
// 				cmds = append(cmds, cmd)
// 			}
// 			app.status, cmd = app.status.Update(msg)
// 			cmds = append(cmds, cmd)

// 			app.state = welcomeView
// 		}

// 		return app, tea.Batch(cmds...)
// 	}

// 	switch msg := msg.(type) {
// 	case tea.KeyMsg:
// 		switch {
// 		// nop, right
// 		case key.Matches(msg, app.keys.Quit):
// 			app.quite <- struct{}{}
// 			return app, tea.Quit
// 		// indicated that the dev wants to switch the view.
// 		// Once this key is hit we need to wait for input
// 		case key.Matches(msg, app.keys.Input) && app.state != initializing && app.state != welcomeView:
// 			app.input.Reset()
// 			app.input.Placeholder = placeholderFormat
// 			app.awaitInput = true
// 			app.input.Focus()
// 			app.input.Prompt = promptFormat
// 			app.requestedCommand = cmdFormat

// 			return app, tea.Batch(cmds...)

// 		case key.Matches(msg, app.keys.Filter) && app.state != initializing && app.state != welcomeView:
// 			app.input.Reset()
// 			app.input.Placeholder = placeholderFilter
// 			app.awaitInput = true
// 			app.input.Focus()
// 			app.input.Prompt = promptFilter
// 			app.requestedCommand = cmdFilter

// 			return app, tea.Batch(cmds...)

// 		// only triggered if input is expected.
// 		// Evaluates the input and coordinates the required
// 		// execution. Current allowed input patterns:
// 		// - [0-9]{1,} -> switch view to formatter with requested input
// 		// - f:[a-zA-Z_-0-9]{1,} -> add filter on log view
// 		case key.Matches(msg, app.keys.Enter) && app.awaitInput:
// 			cmds = append(cmds, app.executeCommand())
// 			app.input.Reset()

// 		// propagate event to formatter and request to format
// 		// previous log line
// 		case key.Matches(msg, app.keys.Up) && !key.Matches(msg, app.keys.Filter) && app.hasInput:
// 			cmds = append(cmds, event.RequestFormatPrevious())
// 			app.ignoreKey = true

// 		// propagate event to formatter and request to format
// 		// next log line
// 		case key.Matches(msg, app.keys.Down) && !key.Matches(msg, app.keys.Filter) && app.hasInput:
// 			cmds = append(cmds, event.RequestFormatNext())
// 			app.ignoreKey = true

// 		// terminate formatting view and propagate event to formatter,
// 		// reinstate tailView as view for app to render
// 		case key.Matches(msg, app.keys.Exit) && app.state != initializing && app.state != welcomeView:
// 			app.awaitInput = false
// 			app.input.Blur()
// 			app.input.Reset()
// 			app.input.Placeholder = placeholderDefault
// 			app.input.Prompt = promptDefault

// 			// unset any filter if some where set
// 			if app.requestedCommand == cmdFilter {
// 				app.buffer.UnsetFilter()
// 				cmds = append(cmds, status.RequestUnFocus())
// 			}

// 			cmds = append(cmds, formatter.RequestQuite(), event.RequestReload())
// 			app.state = tailView
// 		}

// 	case tea.WindowSizeMsg:
// 		app.width = msg.Width
// 		app.height = styles.AvailableHeight(msg.Height)

// 	// event dispatched each time a new stream connects to
// 	// the multiplexer. on-event we need to update the footer
// 	// model with the new stream information as well as update
// 	// the Models state. The Model keeps track of connected beams
// 	// however only cares about the color to use when rendering the logs.
// 	// Model will ensure that the color for the printed logs of a stream
// 	// are matching the color information in the footer
// 	case plexer.Subscriber:
// 		// update max label length for indenting
// 		// while displaying logs
// 		if len(msg) > app.maxLabelLength {
// 			app.maxLabelLength = len(msg)
// 		}

// 		label := string(msg)

// 		if _, ok := app.streams[label]; !ok {
// 			color, _ := styles.RandColor()
// 			app.streams[label] = color
// 		}

// 		app.status, _ = app.status.Update(status.Connection{
// 			Label: label,
// 			Color: app.streams[label],
// 		})
// 		cmds = append(cmds, app.consumeSubscriber)

// 		return app, tea.Batch(cmds...)

// 	// event dispatched each time a beam disconnects from scotty.
// 	// The message itself is the label of the stream which
// 	// disconnected. On a disconnect we need to recompute the
// 	// length of the longest stream label in order to maintain
// 	// pretty indention for logging the logs with the label prefix
// 	case plexer.Unsubscribe:
// 		// we only need to reassign the max value
// 		// if the current max is disconnecting
// 		if len(msg) >= app.maxLabelLength {
// 			max := 0
// 			for label := range app.streams {
// 				if len(label) > max && label != string(msg) {
// 					max = len(label)
// 				}
// 			}
// 			app.maxLabelLength = max
// 		}
// 		cmds = append(cmds, app.consumerUnsubscribe)

// 	// event dispatched by the multiplexer each time a client/stream
// 	// sends a log linen.
// 	// The App needs to add the ansi color code stored for the stream
// 	// to the dispatched message before adding the data to the ring buffer.
// 	// Further processing happening in active view (views[state]).
// 	// For references see:
// 	// - pager.Update()
// 	case plexer.Message:
// 		// in all cases the first view we show whenever
// 		// the first message is received is the pager
// 		if app.state == welcomeView {
// 			app.state = tailView
// 		}

// 		color := app.streams[msg.Label]

// 		space := app.maxLabelLength - len(msg.Label)
// 		if space < 0 {
// 			space = 0
// 		}

// 		prefix := lipgloss.NewStyle().
// 			Foreground(color).
// 			Render(
// 				msg.Label+strings.Repeat(" ", space),
// 			) + " | "

// 		app.logstore.Insert(msg.Label, len(prefix), append([]byte(prefix), bytes.TrimSpace(msg.Data)...))
// 		cmds = append(cmds, app.consumeMsg)

// 	case plexer.Error:
// 		// good question guess pipe to status???
// 	}

// 	// update current model
// 	app.views[app.state], cmd = app.views[app.state].Update(msg)
// 	cmds = append(cmds, cmd)

// 	app.status, cmd = app.status.Update(msg)
// 	cmds = append(cmds, cmd)

// 	if !app.ignoreKey {
// 		app.input, cmd = app.input.Update(msg)
// 		cmds = append(cmds, cmd)
// 	}
// 	app.ignoreKey = false

// 	return app, tea.Batch(cmds...)
// }

// func (app *App) View() string {
// 	if app.state == initializing || app.state == welcomeView {
// 		return app.views[welcomeView].View()
// 	}

// 	return lipgloss.JoinVertical(lipgloss.Left,
// 		app.status.View(),
// 		app.views[app.state].View(),
// 		app.input.View(),
// 	)
// }

// // executeCommands interprets the users input
// // and based on the previous keystroke a follow-up
// // action in form of an tea.Cmd is evaluated.
// func (app *App) executeCommand() tea.Cmd {

// 	value := app.input.Value()

// 	switch app.requestedCommand {
// 	case cmdFormat:
// 		index, err := strconv.Atoi(value)
// 		if err != nil {
// 			debug.Print("unable to parse index: %q: %v", value, err)
// 		}
// 		app.hasInput = true

// 		// app.state = formatView
// 		return event.RequestFormatInit(index)

// 	case cmdFilter:

// 		const (
// 			overwrite = iota
// 			add
// 			remove
// 		)

// 		var streams []string
// 		var operation = overwrite

// 		if strings.HasPrefix(value, "+") {
// 			operation = add
// 		}

// 		if strings.HasPrefix(value, "-") {
// 			operation = remove
// 		}

// 		switch operation {
// 		case overwrite:
// 			for _, s := range strings.Split(value, ",") {
// 				streams = append(streams, strings.TrimSpace(s))
// 			}
// 			app.buffer.ApplyFilter(func(item, label string, data []byte) bool {
// 				return item == label
// 			}, streams...)

// 			return tea.Batch(status.RequestFocus(streams...), event.RequestReload())
// 		case add:
// 			for _, s := range strings.Split(value[1:], ",") {
// 				streams = append(streams, strings.TrimSpace(s))
// 			}
// 			app.buffer.AddFilter(streams...)

// 			return tea.Batch(status.RequestAddFocus(streams...), event.RequestReload())

// 		// remove op can only remove single items at a time
// 		case remove:
// 			stream := strings.TrimSpace(value[1:])
// 			app.buffer.RemoveFilter(stream)

// 			return tea.Batch(status.RequestRemoveFocus(stream), event.RequestReload())
// 		}

// 	}
// 	return nil
// }
