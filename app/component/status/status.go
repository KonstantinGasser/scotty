package status

import (
	"fmt"
	"sync"

	"github.com/KonstantinGasser/scotty/app/styles"
	plexer "github.com/KonstantinGasser/scotty/multiplexer"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/exp/slices"
)

var (
	ModelStyle = lipgloss.NewStyle().
			Padding(0, 1).
			Border(lipgloss.DoubleBorder(), false, false, true, false)

	newNotificationStyle = lipgloss.NewStyle().
				MarginLeft(1).
				Foreground(styles.DefaultColor.Border).
				Render(labelNewNotification)

	spacing = styles.Spacer(1).Render("")

	filterStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("7"))
)

const (
	labelDisconnected    = "SIGINT"
	labelNewNotification = "(+1)"
)

type Connection struct {
	Label string
	Color lipgloss.Color
}

type stream struct {
	label        string
	colorBg      lipgloss.Color
	colorFg      lipgloss.Color
	style        lipgloss.Style
	count        int
	disconnected bool
	focused      bool

	// works in conjunction with isFormatMode
	// and is flipped when the parent mode
	// propagates a multiplexer.Message msg to the model.
	// Only if both isFormatMode and hasNewMessage are true
	// an indication for the user about new (hidden) messages
	// must be shown
	hasNewMessages bool
}

type Model struct {
	ready         bool
	width, height int

	// any error happing anywhere
	// in the application should be shown
	// in the Model.
	// err represents the latest error
	err error

	// guards below fields
	mtx sync.RWMutex
	// number of logs stream by all streams
	// dropping a stream results in logCount - len(stream)
	logCount int
	// slice of beams which are currently connected to scotty
	connectedBeams map[string]*stream

	// streamIndex holds the index of the stream
	// to allow O(1) indexing of the streams slice.
	// Problem we are trying to solve with this is the
	// fact that maps don't guaranty same ordering when
	// looping over the map. However, this leeds to the
	// Model switching the stream info boxes which is annoying...
	streamIndex map[string]int
	// streams is the slice with the actual information about a stream.
	// using the streamsIndex we can do a O(1) lookup for a specific stream
	// but maintain ordering when looping over the list when calling View()
	streams []stream

	// isFormatMode is indicated by the Models parent model.
	// It tells whether or not the user is currently browsing
	// through the logs.
	// If true and a new message is received by a stream we want
	// to add an indicated next to the stream info telling that new
	// logs are available
	isFormatMode bool

	// hasFilter indicates that one or more streams are focused.
	hasFilter bool
}

func New() *Model {
	return &Model{
		ready: false,
		err:   nil,

		mtx:            sync.RWMutex{},
		logCount:       0,
		connectedBeams: map[string]*stream{},

		streamIndex:  make(map[string]int),
		streams:      make([]stream, 0, 4), // 4 is a wage assumption of the number of potential streams connecting to scotty. Could be me could be less, but might help with not moving stuff around so much?
		isFormatMode: false,
	}
}

func (model *Model) Init() tea.Cmd {
	return nil
}

func (model *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	var (
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if !model.ready {
			model.ready = true
		}
		model.width = msg.Width
		model.height = styles.AvailableHeight(msg.Height)

		// model.height = msg.Height -> not really interested in the tty height
		return model, nil

	case requestedUnFocus:
		model.hasFilter = false
		for i := range model.streams {
			model.streams[i].focused = false
		}
	// a filter was applied we want to highlight the color
	// of the requested stream while reducing the others
	case requestedFocus:
		model.hasFilter = true

		for key, index := range model.streamIndex {

			if ok := slices.Contains(msg, key); !ok {
				model.streams[index].focused = false
				continue
			}

			model.streams[index].focused = true
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			model.isFormatMode = true

		// invoked by parent model when leaving
		// formatting mode thus we want to reset
		// values used to show hidden notifications
		case "esc", "q":
			model.isFormatMode = false
			for i := range model.streams {
				model.streams[i].hasNewMessages = false
			}
		}
	// whenever a stream connects to scotty the event
	// is propagated. The Model uses the event to
	// display connected stream and the number of logs streamed.
	// Streams which disconnect temporally are removed from the
	// Model state (color coding is delegated from the pager and thus
	// stays the same unless the same stream connects with a different label).
	// Each stream has a random background color for readability the Model
	// either uses a white or black foreground color
	case Connection:
		fg := styles.InverseColor(msg.Color)
		index, ok := model.streamIndex[msg.Label]
		if !ok {
			model.streams = append(model.streams, stream{
				label:   msg.Label,
				colorBg: msg.Color,
				colorFg: fg,
				style: lipgloss.NewStyle().
					Background(msg.Color).
					Foreground(fg).
					Padding(0, 1).
					Bold(true),
				count: 0,
			})
			// don't forget to update the index map
			model.streamIndex[msg.Label] = len(model.streams) - 1
			break
		}

		model.streams[index].disconnected = false

	case plexer.Unsubscribe:
		index := model.streamIndex[string(msg)]
		model.streams[index].disconnected = true

	case plexer.Error:
		// QUESTION @KonstantinGasser:
		// How do I unset the error say after 15 seconds?
		model.err = msg
	case plexer.Message:
		// lookup the stream which dispatched the event
		// and increase the log count
		index, ok := model.streamIndex[msg.Label]
		if !ok {
			break
		}

		model.streams[index].count++

		if model.isFormatMode && !model.streams[index].hasNewMessages && !model.streams[index].focused {
			model.streams[index].hasNewMessages = true
		}
	}

	return model, tea.Batch(cmds...)
}

func (model *Model) View() string {

	if !model.ready {
		return "initializing..."
	}

	var items = []string{}
	var filtered = []string{}

	if len(model.streams) <= 0 && model.err == nil {
		txt := "beam the logs up, scotty is ready"
		items = append(items,
			styles.StatusBarLogCount(txt),
		)
	}

	var padding string = spacing
	var info string
	for i, stream := range model.streams {

		if i >= len(model.streams)-1 {
			padding = ""
		}

		// not space after last one thou
		info = fmt.Sprint(stream.count)

		if model.hasFilter {

			if stream.hasNewMessages {
				info += newNotificationStyle
			}

			if stream.disconnected {
				info += labelDisconnected
			}

			if !stream.focused {
				items = append(items, stream.style.Copy().
					Background(lipgloss.Color("#c0c0c0")).
					Foreground(lipgloss.Color("#808080")).
					Render(stream.label+":"+fmt.Sprint(info)),
					padding,
				)
				continue
			}

			items = append(items, stream.style.
				Render(stream.label+":"+fmt.Sprint(info)),
				padding,
			)
			filtered = append(filtered, stream.label)
			continue
		}

		if stream.disconnected {
			items = append(items, stream.style.Render(
				stream.label+":"+fmt.Sprint(labelDisconnected),
			),
				padding,
			)
			continue
		}

		if stream.hasNewMessages {
			items = append(items, stream.style.Render(
				stream.label+":"+fmt.Sprint(info)+newNotificationStyle,
			),
				padding,
			)
			continue
		}

		items = append(items, stream.style.Render(
			stream.label+":"+fmt.Sprint(stream.count),
		),
			padding,
		)
	}

	if model.err != nil {
		items = append(items,
			styles.Spacer(2).Render("	"), // add some space next to the beams
			styles.ErrorInfo(model.err.Error()),
		)
	}

	streams := lipgloss.NewStyle().Render(
		lipgloss.JoinHorizontal(lipgloss.Top,
			items...,
		),
	)

	// display filter to the right of the status bar
	filter := "filters: non set"
	if len(filtered) > 0 {
		filter = filterStyle.Render(fmt.Sprintf("filters: %v", filtered))
	}

	// add space between streams and filter to push the filter
	// to the right of the status bar
	items = append(items,
		styles.Spacer(
			model.width-lipgloss.Width(streams)-lipgloss.Width(filter)-10,
		).Render(""),
	)

	items = append(items, filter)

	return ModelStyle.
		Padding(0, 1).
		Width(model.width - 1). // account for padding left/right
		Render(
			lipgloss.JoinHorizontal(lipgloss.Top,
				items...,
			),
		)
}

type requestedFocus []string

func RequestFocus(streams ...string) tea.Cmd {
	return func() tea.Msg {
		return requestedFocus(streams)
	}
}

type requestedUnFocus struct{}

func RequestUnFocus() tea.Cmd {
	return func() tea.Msg {
		return requestedUnFocus{}
	}
}
