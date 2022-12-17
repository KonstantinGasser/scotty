package base

import (
	"fmt"
	"strings"

	"github.com/KonstantinGasser/scotty/streams"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

type Streamer interface {
	Stream() <-chan streams.Message
}

type Model struct {
	// quite channel indicated that the user/tea received
	// a quite (q) or ctl+c
	quite chan<- struct{}
	// raw messages coming from each stream can be received here
	// where each Message includes the context of the stream such as
	// the label of the stream
	messages <-chan streams.Message
	errors   <-chan error
	logs     []string
	view     viewport.Model
}

func New(quite chan<- struct{}, errors <-chan error, messages <-chan streams.Message) *Model {

	return &Model{
		quite:    quite,
		messages: messages,
		errors:   errors,
		view:     viewport.New(60, 30),
	}
}

func (m *Model) Init() tea.Cmd {
	return m.wait
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quite <- struct{}{}
			return m, tea.Quit
		}
	case streams.Message:
		m.logs = append(m.logs, fmt.Sprintf("[%s] %s", msg.Label, msg.Raw))

		var builder = strings.Builder{}

		for _, log := range m.logs {
			builder.WriteString(log)
		}
		m.view.SetContent(builder.String())
		return m, m.wait
	case error:
		m.view.SetContent(msg.Error())
		return m, m.wait
	}

	return m, nil
}

func (m *Model) View() string {
	return m.view.View()
}

// waits and blocks until a new event is received in order
// to return a bubbletea understandable tea.Cmd
func (m *Model) wait() tea.Msg {
	select {
	case err := <-m.errors:
		return err
	case msg := <-m.messages:
		return msg
	}
}
