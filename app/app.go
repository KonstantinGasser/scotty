package app

import (
	"time"

	"github.com/KonstantinGasser/scotty/app/bindings"
	"github.com/KonstantinGasser/scotty/app/component/browsing"
	"github.com/KonstantinGasser/scotty/app/component/docs"
	"github.com/KonstantinGasser/scotty/app/component/info"
	"github.com/KonstantinGasser/scotty/app/component/querying"
	"github.com/KonstantinGasser/scotty/app/component/tailing"
	"github.com/KonstantinGasser/scotty/app/component/welcome"
	"github.com/KonstantinGasser/scotty/app/styles"
	"github.com/KonstantinGasser/scotty/debug"
	"github.com/KonstantinGasser/scotty/store"
	"github.com/KonstantinGasser/scotty/stream"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	tabUnset = iota - 1
	tabFollow
	tabBrowse
	tabQuery
	tabDocs
)

const (
	scopeFollow = "SCOPE-FOLLOW"
	scopeBrowse = "SCOPE-BROWSE"
)

type mode struct {
	label string
	bg    lipgloss.Color
}

var (
	modeFollowing mode = mode{label: "FOLLOWING", bg: lipgloss.Color("#98c379")}
	modeBrowsing  mode = mode{label: "BROWSING", bg: lipgloss.Color("#98c378")}
	modePaused    mode = mode{label: "PAUSED", bg: lipgloss.Color("#ff9640")}
	modeGlobalCmd mode = mode{label: "GLOBAL (esc for exit)", bg: lipgloss.Color("54")}

	globalKey = key.NewBinding(key.WithKeys(" "))
)

type streamConfig struct {
	color lipgloss.Color
}

type App struct {
	/* internal properties */
	// indication to close and stop work.
	// Signal send from outside
	quit chan<- struct{}
	// availabel dimension
	ttyWidth, ttyHeight int
	grid                styles.Grid
	// false until initial tea.WindowSizeMsg send
	// and App is initialized
	ready bool
	// key bindings
	bindings       *bindings.Map
	ignoreBindings []key.Binding
	/* stream / i/o properties */
	// channels to consume stream events
	consumer   stream.Consumer
	subscriber map[string]streamConfig

	// place where all logs are written
	// to. App manly uses it for inserts
	logstore *store.Store

	/* layout properties */
	// finished parsed and build tabs
	// where one tab is shown as active.
	// Default is welcome-tab:active
	headerComponent *styles.Tabs
	// indication which tab is currently
	// active and thereby which component
	activeTab int

	/* component specific properties */
	footerComponent tea.Model
	// map of all available components mapped to
	// the available tabs
	components map[int]tea.Model
}

func New(q chan<- struct{}, refresh time.Duration, lStore *store.Store, consumer stream.Consumer) *App {

	app := &App{
		quit:      q,
		ttyWidth:  -1, // unset/invalid
		ttyHeight: -1, // unset/invalid
		ready:     false,
		bindings:  bindings.NewMap(),

		consumer:   consumer,
		subscriber: make(map[string]streamConfig),
		logstore:   lStore,

		headerComponent: styles.NewTabs(0, "follow logs", "browse logs", "query logs", "docs"),
		activeTab:       tabUnset,

		footerComponent: info.New(),
		components: map[int]tea.Model{
			tabFollow: tailing.New(lStore.NewPager(0, 0, refresh)),
			tabBrowse: browsing.New(lStore.NewFormatter(0, 0)),
			tabQuery:  querying.New(),
			tabDocs:   docs.New(),
		},
	}

	app.bindings.Bind("ctrl+c").Action(func(msg tea.KeyMsg) tea.Cmd {
		app.quit <- struct{}{}
		return tea.Quit
	})

	app.bindings.Bind(" ").
		Action(func(msg tea.KeyMsg) tea.Cmd {
			return info.RequestMode(modeGlobalCmd.label, modeGlobalCmd.bg)
		}).
		Option("f").
		Action(func(msg tea.KeyMsg) tea.Cmd {
			if app.activeTab == tabFollow {
				return nil
			}
			app.activeTab = tabFollow
			app.headerComponent.SetActive(app.activeTab)

			return info.RequestMode(modeFollowing.label, modeFollowing.bg)
		})

	app.bindings.Bind(" ").OnESC(func(msg tea.KeyMsg) tea.Cmd {
		return info.RequestMode(modeBrowsing.label, modeBrowsing.bg)
	}).
		Option("b").Action(func(msg tea.KeyMsg) tea.Cmd {
		if app.activeTab == tabBrowse {
			return nil
		}

		app.activeTab = tabBrowse
		app.headerComponent.SetActive(app.activeTab)
		return nil
	})

	app.bindings.Bind(" ").Option("s").Action(func(msg tea.KeyMsg) tea.Cmd {
		if app.activeTab == tabQuery {
			return nil
		}

		app.activeTab = tabQuery
		app.headerComponent.SetActive(app.activeTab)
		return nil
	})

	app.bindings.Bind(" ").Option("d").Action(func(msg tea.KeyMsg) tea.Cmd {
		if app.activeTab == tabDocs {
			return nil
		}

		app.activeTab = tabDocs
		app.headerComponent.SetActive(app.activeTab)
		return nil
	})

	app.bindings.Debug()

	return app
}

func (app App) Init() tea.Cmd {
	return tea.Batch(
		app.consumeMsg,
		app.consumeSubscriber,
		app.consumeUnsubscribe,
		app.consumeErrs,
	)
}

func (app *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	var (
		cmds []tea.Cmd
		cmd  tea.Cmd
	)

	debug.Print("[App:Update] Msg: %T\n", msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// we dont allow any key binding actions
		// while no beam has connected and we are in the
		// welcome screen
		if app.activeTab == tabUnset {
			if key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+c"))) {
				cmds = append(cmds, app.bindings.Exec(msg).Call(msg))
				return app, tea.Batch(cmds...)
			}
			break
		}

		if !app.bindings.Matches(msg) {
			// does not mean the action component
			// might not do something with the event
			// so we pass it down to the active component
			app.components[app.activeTab], cmd = app.components[app.activeTab].Update(msg)
			cmds = append(cmds, cmd)
			return app, tea.Batch(cmds...)
		}

		cmds = append(cmds, app.bindings.Exec(msg).Call(msg))
		return app, tea.Batch(cmds...)

	case tea.WindowSizeMsg:

		// iterate over all components as they are not
		// aware of the inital width and height of the tty
		if !app.ready {
			app.grid = styles.NewGrid(msg.Width, msg.Height)
			app.ready = true
		} else {
			app.grid.Adjust(msg.Width, msg.Height)
		}

		app.footerComponent, cmd = app.footerComponent.Update(app.grid.FooterLine.Dims())
		cmds = append(cmds, cmd)
		for i, comp := range app.components {
			app.components[i], cmd = comp.Update(app.grid.Content.Dims())
			cmds = append(cmds, cmd)
		}

		return app, tea.Batch(cmds...)

	// triggered by the tailing component indicating that the "p" key was pressed
	// and the pager stops at the current index only updating in the background
	case tailing.PauseRequest:
		cmds = append(cmds,
			info.RequestPause(),
			info.RequestMode(modePaused.label, modePaused.bg),
		)

	// triggered by the tailing component indicating that the pager resumes to render
	// the latest logs
	case tailing.ResumeRequest:
		cmds = append(cmds,
			info.RequestResume(),
			info.RequestMode(modeFollowing.label, modeFollowing.bg),
		)

	// triggered each time a new stream connects successfully to scotty and is procssed
	// by the stream. A random color is assigned to the stream if not yet pressent
	// (identified by its label). An update about the new stream is propagated to the info
	// component.
	case stream.Subscriber:
		if _, ok := app.subscriber[string(msg)]; !ok {
			fg, _ := styles.RandColor()
			app.subscriber[string(msg)] = streamConfig{color: fg}
		}

		app.footerComponent, _ = app.footerComponent.Update(
			info.RequestSubscribe(string(msg), app.subscriber[string(msg)].color)(),
		)

		cmds = append(cmds, app.consumeSubscriber)
		return app, tea.Batch(cmds...)

	case stream.Unsubscribe:
		if app.activeTab == tabFollow {
			app.components[tabFollow], _ = app.components[tabFollow].Update(tailing.RequestRefresh()())
		}
		app.footerComponent, _ = app.footerComponent.Update(info.RequestUnsubscribe(string(msg))())

		cmds = append(cmds, app.consumeUnsubscribe)
		return app, tea.Batch(cmds...)

	// triggered each time a new message is pushed from the stream to
	// the consumer.
	// Requires to identify the stream the message is from, build the prefix
	// and to store the message in the log-store. Furthermore, inserts into
	// the log-store will happend dispite the active tab. This allows background
	// updates of the follow-components between tab switches.
	case stream.Message:
		if app.activeTab == tabUnset {
			app.activeTab = tabFollow
			cmds = append(cmds, info.RequestMode(modeFollowing.label, modeFollowing.bg))
			// no longer required. Default of active Tab is zero
			// and tabFollow == 0
			// ...
			// app.updateActiveTab()
		}

		config, ok := app.subscriber[msg.Label]
		if !ok {
			break
		}

		prefix := lipgloss.NewStyle().Foreground(config.color).Render(msg.Label) + " | "

		app.logstore.Insert(msg.Label, len(prefix), append([]byte(prefix), msg.Data...))
		// update follow component asap in order to allow background updates while
		// in a different tab
		app.components[tabFollow], _ = app.components[tabFollow].Update(msg)
		cmds = append(cmds, app.consumeMsg)

		app.footerComponent, _ = app.footerComponent.Update(info.RequestIncrement(msg.Label)())
		return app, tea.Batch(cmds...)
	}

	// follow component is updates asap after a message is received
	// if app.activeTab != tabUnset && app.activeTab != tabFollow {
	// 	app.components[app.activeTab], cmd = app.components[app.activeTab].Update(msg)
	// 	cmds = append(cmds, cmd)
	// }

	app.footerComponent, cmd = app.footerComponent.Update(msg)
	cmds = append(cmds, cmd)

	return app, tea.Batch(cmds...)
}

func (app App) View() string {

	if app.activeTab == tabUnset {
		return welcome.New(app.grid.FullWidth, app.grid.FullHeight).View()
	}

	return lipgloss.NewStyle().
		// Padding(styles.ContentPaddingVertical, 0).
		Render(
			lipgloss.JoinVertical(lipgloss.Left,
				app.headerComponent.View(),
				app.components[app.activeTab].View(),
				app.grid.FooterLine.Render(app.footerComponent.View()),
			),
		)
}

/* consume* yields back a tea.Msg piped through a channel ending in the app.Update func */
func (app *App) consumeMsg() tea.Msg         { return <-app.consumer.Messages() }
func (app *App) consumeErrs() tea.Msg        { return <-app.consumer.Errors() }
func (app *App) consumeSubscriber() tea.Msg  { return <-app.consumer.Subscribers() }
func (app *App) consumeUnsubscribe() tea.Msg { return <-app.consumer.Unsubscribers() }
