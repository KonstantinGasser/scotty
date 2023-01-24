package pager

import (
	"fmt"
	"sort"
	"sync"

	"github.com/KonstantinGasser/scotty/app/styles"
	plexer "github.com/KonstantinGasser/scotty/multiplexer"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	footerStyle = lipgloss.NewStyle().
			Margin(0, 2)

	beamSpacer = styles.Spacer(1).Render("")
)

type stream struct {
	colorBg lipgloss.Color
	colorFg lipgloss.Color
	style   func(string) string
	count   int
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
}

func newFooter(w, h int) *footer {
	return &footer{
		width:  w,
		height: h,
		err:    nil,

		mtx:            sync.RWMutex{},
		logCount:       0,
		connectedBeams: map[string]*stream{},
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
		f.width = msg.Width - 2 // account for margin
		f.height = msg.Height
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
		f.connectedBeams[msg.label] = &stream{
			colorBg: msg.color,
			colorFg: fg,
			style: lipgloss.NewStyle().
				Background(msg.color).
				Foreground(fg).
				Padding(0, 1).
				Bold(true).
				Render,
			count: 0,
		}

	case plexer.Error:
		// QUESTION @KonstantinGasser:
		// How do I unset the error say after 15 seconds?
		f.err = msg
	case plexer.Message:
		// plexer.BeamMessage needs to be extended with
		// information about the stream such as the label of it
		// only then we can increase the respective count
		f.connectedBeams[msg.Label].count++
	}

	return f, tea.Batch(cmds...)
}

func (f *footer) View() string {

	var items = []string{}

	if len(f.connectedBeams) <= 0 {
		txt := "beam the logs up, scotty is ready"
		items = append(items,
			styles.StatusBarLogCount(txt),
		)
	}

	// add a little space between beam labels
	var i int
	var labels []string
	for label, info := range f.connectedBeams {
		if i < len(f.connectedBeams) {
			labels = append(items, beamSpacer, info.style(
				label+":"+fmt.Sprint(info.count),
			))
			i++
			continue
		}
		// not space after last one thou
		labels = append(items, info.style(
			label+":"+fmt.Sprint(info.count),
		))
		i++
	}

	// since maps are not ordered the ui renders services
	// with changing order which is enjoying to see
	// as such order the labels by name
	// so what should we do?
	// sorting the slice of styled strings?
	// mhm not a fan of it..
	// maybe some other format would work better..
	// maybe for each stream we can accept O(n) tc
	// when checking if a stream is already present
	// ...
	sort.Strings(labels)
	items = append(items, labels...)

	if f.err != nil {
		items = append(items,
			styles.Spacer(2).Render(""), // add some space next to the beams
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
