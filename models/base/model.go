package base

import (
	"fmt"
	"strings"

	"github.com/KonstantinGasser/scotty/streams"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Streamer interface {
	Stream() <-chan streams.Message
}

type Model struct {
	help     help.Model
	bindings keyMap

	// index is the current index of the log which is selected/focused
	index int

	// quite channel indicated that the user/tea received
	// a quite (q) or ctl+c
	quite chan<- struct{}
	// raw messages coming from each stream can be received here
	// where each Message includes the context of the stream such as
	// the label of the stream
	messages    <-chan streams.Message
	logs        []string // will be replaced by a ring buffer
	subscriber  <-chan string
	subscribers []string
	errors      <-chan error
	view        viewport.Model
}

func New(quite chan<- struct{}, errors <-chan error, subs <-chan string, messages <-chan streams.Message) (*Model, error) {
	const width = 40

	vp := viewport.New(width, 50)

	vp.Style = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		PaddingRight(2)

	vp.SetContent("no logs received so far")

	return &Model{
		help:        help.New(),
		bindings:    keys,
		quite:       quite,
		messages:    messages,
		subscriber:  subs,
		subscribers: make([]string, 0),
		errors:      errors,
		view:        vp,
	}, nil
}

func (m *Model) Init() tea.Cmd {
	return m.wait
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.resolveBinding(msg)
	case error:
		m.view.SetContent(msg.Error())
		return m, m.wait
	case tea.WindowSizeMsg:
		// If we set a width on the help menu it can it can gracefully truncate
		// its view as needed.
		m.help.Width = msg.Width

	case streams.Message:
		m.logs = append(m.logs, fmt.Sprintf("[%s] %s", msg.Label, msg.Raw))

		var builder = strings.Builder{}

		for _, log := range m.logs {
			builder.WriteString(log)
		}
		m.view.Update(builder.String())
		return m, m.wait
	}

	return m, nil
}

func (m *Model) View() string {
	helpView := m.help.View(m.bindings)

	return m.view.View() + "\n" + styleHelp(helpView)
}

// waits and blocks until a new event is received in order
// to return a bubbletea understandable tea.Cmd
func (m *Model) wait() tea.Msg {
	select {
	case err := <-m.errors:
		return err
	case msg := <-m.messages:
		return msg
		// case sub := <-m.subscriber:
		// 	return sub
	}
}
