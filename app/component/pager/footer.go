package pager

import (
	"fmt"
	"sync"

	"github.com/KonstantinGasser/scotty/app/styles"
	plexer "github.com/KonstantinGasser/scotty/multiplexer"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	footerStyle = lipgloss.NewStyle().Padding(0, 1)

	spacing = styles.Spacer(1).Render("")
)

const (
	labelDisconnected = "SIGINT"
)

type stream struct {
	label        string
	colorBg      lipgloss.Color
	colorFg      lipgloss.Color
	style        func(string) string
	count        int
	disconnected bool
}

type footer struct {
	width, height int

	// any error happing anywhere
	// in the application should be shown
	// in the footer.
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
	// footer switching the stream info boxes which is annoying...
	streamIndex map[string]int
	// streams is the slice with the actual information about a stream.
	// using the streamsIndex we can do a O(1) lookup for a specific stream
	// but maintain ordering when looping over the list when calling View()
	streams []stream
}

func newFooter(w, h int) *footer {
	return &footer{
		width: w,
		err:   nil,

		mtx:            sync.RWMutex{},
		logCount:       0,
		connectedBeams: map[string]*stream{},

		streamIndex: make(map[string]int),
		streams:     make([]stream, 0, 4), // 4 is a wage assumption of the number of potential streams connecting to scotty. Could be me could be less, but might help with not moving stuff around so much?
	}
}

func (f *footer) Init() tea.Cmd {
	return nil
}

func (f *footer) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	var (
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		f.width = msg.Width
		// f.height = msg.Height -> not really interested in the tty height
		return f, nil

	// whenever a stream connects to scotty the event
	// is propagated. The footer uses the event to
	// display connected stream and the number of logs streamed.
	// Streams which disconnect temporally are removed from the
	// footer state (color coding is delegated from the pager and thus
	// stays the same unless the same stream connects with a different label).
	// Each stream has a random background color for readability the footer
	// either uses a white or black foreground color
	case subscriber:
		fg := styles.InverseColor(msg.color)
		index, ok := f.streamIndex[msg.label]
		if !ok {
			f.streams = append(f.streams, stream{
				label:   msg.label,
				colorBg: msg.color,
				colorFg: fg,
				style: lipgloss.NewStyle().
					Background(msg.color).
					Foreground(fg).
					Padding(0, 1).
					Bold(true).
					Render,
				count: 0,
			})
			// don't forget to update the index map
			f.streamIndex[msg.label] = len(f.streams) - 1
			break
		}

		f.streams[index].disconnected = false

	case plexer.Unsubscribe:
		index := f.streamIndex[string(msg)]
		f.streams[index].disconnected = true

	case plexer.Error:
		// QUESTION @KonstantinGasser:
		// How do I unset the error say after 15 seconds?
		f.err = msg
	case plexer.Message:
		// lookup the stream which dispatched the event
		// and increase the log count
		index, ok := f.streamIndex[msg.Label]
		if !ok {
			break
		}

		f.streams[index].count++
	}

	return f, tea.Batch(cmds...)
}

func (f *footer) View() string {

	var items = []string{}

	if len(f.streams) <= 0 && f.err == nil {
		txt := "beam the logs up, scotty is ready"
		items = append(items,
			styles.StatusBarLogCount(txt),
		)
	}

	for i, stream := range f.streams {
		if i >= len(f.streams)-1 {
			// not space after last one thou

			var info any = stream.count
			if stream.disconnected {
				info = labelDisconnected
			}

			items = append(items, stream.style(
				stream.label+":"+fmt.Sprint(info),
			))
			continue
		}

		items = append(items, stream.style(
			stream.label+":"+fmt.Sprint(stream.count),
		),
			spacing,
		)
	}

	if f.err != nil {
		items = append(items,
			styles.Spacer(2).Render("	"), // add some space next to the beams
			styles.ErrorInfo(f.err.Error()),
		)
	}

	return footerStyle.
		Render(
			lipgloss.JoinHorizontal(lipgloss.Left,
				items...,
			),
		)
}
