package app

import (
	"bytes"
	"strconv"
	"strings"

	"github.com/KonstantinGasser/scotty/app/component/browsing"
	"github.com/KonstantinGasser/scotty/app/component/docs"
	"github.com/KonstantinGasser/scotty/app/component/info"
	"github.com/KonstantinGasser/scotty/app/component/querying"
	"github.com/KonstantinGasser/scotty/app/component/tailing"
	"github.com/KonstantinGasser/scotty/app/component/welcome"
	"github.com/KonstantinGasser/scotty/app/event"
	"github.com/KonstantinGasser/scotty/app/styles"
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
	bindings       bindings
	ignoreBindings []key.Binding
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
			tabBrowse: browsing.New(lStore.NewFormatter(0, 0)),
			tabQuery:  querying.New(),
			tabDocs:   docs.New(),
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
		// some compontens requested to ignore these keys as they are relevent to be
		// processed within the component itself
		case key.Matches(msg, app.ignoreBindings...):
			// propagate ignored keys to the active componten
			app.compontens[app.activeTab], cmd = app.compontens[app.activeTab].Update(msg)
			cmds = append(cmds, cmd)
			return app, tea.Batch(cmds...)
		case key.Matches(msg, app.bindings.SwitchTab):
			tabIndex, _ := strconv.ParseInt(msg.String(), 10, 64)
			tabIndex = tabIndex - 1 // -1 as it is displayed as 1 2 3 4 but index at 0

			if app.activeTab == int(tabIndex) {
				return app, tea.Batch(cmds...)
			}

			app.activeTab = int(tabIndex)
			app.updateActiveTab()

		}
		if app.activeTab > tabUnset {
			app.compontens[app.activeTab], cmd = app.compontens[app.activeTab].Update(msg)
			cmds = append(cmds, cmd)
		}

		return app, tea.Batch(cmds...)
	case event.BlockKeys:
		app.ignoreBindings = append(app.ignoreBindings, key.NewBinding(key.WithKeys(msg...)))
	case event.ReleaseKeys:
		app.ignoreBindings = nil
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
		return app, tea.Batch(cmds...)
	// triggered each time a new stream connects successfully to scotty and is procssed
	// by the multiplexer. A random color is assigned to the stream if not yet pressent
	// (identified by its label). An update about the new stream is propagated to the info
	// component.
	case multiplexer.Subscriber:
		if _, ok := app.streamConfigs[string(msg)]; !ok {
			fg, _ := styles.RandColor()
			app.streamConfigs[string(msg)] = streamConfig{color: fg}
		}

		app.infoComponent, _ = app.infoComponent.Update(
			info.NewBeam(string(msg), app.streamConfigs[string(msg)].color),
		)

		cmds = append(cmds, app.consumeSubscriber)
		return app, tea.Batch(cmds...)

	case multiplexer.Unsubscribe:
		app.infoComponent, _ = app.infoComponent.Update(info.DisconnectBeam(msg))
		return app, tea.Batch(cmds...)

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

		app.infoComponent, _ = app.infoComponent.Update(event.Increment(msg.Label))
		return app, tea.Batch(cmds...)
	}

	// follow component is updates asap after a message is received
	if app.activeTab != tabUnset && app.activeTab != tabFollow {
		app.compontens[app.activeTab], cmd = app.compontens[app.activeTab].Update(msg)
		cmds = append(cmds, cmd)
	}

	app.infoComponent, cmd = app.infoComponent.Update(msg)
	cmds = append(cmds, cmd)

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
