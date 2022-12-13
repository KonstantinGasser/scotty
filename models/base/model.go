package base

import (
	"bufio"
	"fmt"
	"io"
	"net"

	tea "github.com/charmbracelet/bubbletea"
)

type Event interface {
	// analog to tea.Model.View
	// what ever the event returns here as a string
	// will be rendered
	View() string
	// Index should return the index at which the
	// event is stored (see ring/buffer.go)
	Index() uint8
}

type Model struct {
	quite  chan<- struct{}
	stream <-chan net.Conn
	events chan Event
	latest Event
}

func New(quite chan<- struct{}, stream <-chan net.Conn, evts chan Event) *Model {
	return &Model{
		quite:  quite,
		stream: stream,
		events: evts,
	}
}

type SomeEvent struct {
	msg string
}

func (se SomeEvent) View() string { return se.msg }

func (se SomeEvent) Index() uint8 { return 0 }

func receiveEvents(m *Model) tea.Cmd {
	return func() tea.Msg {
		for conn := range m.stream {
			go func(c net.Conn) {

				reader := bufio.NewReader(c)
				for {
					line, err := reader.ReadString('\n')
					if err != nil {
						if err == io.EOF {
							// what should we do here? good questions
							// need to build further logic to make sense
							// out of it
							return
						}
						fmt.Printf("unable to read from stream: %v", err)
						return
					}
					m.events <- SomeEvent{msg: line}
				}
			}(conn)
		}

		return "Done" // has no effect and is just here to satisfy the compiler
	}
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		receiveEvents(m),
		wait(m.events),
	)
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quite <- struct{}{}
			return m, tea.Quit
		}
	case Event:
		m.latest = msg
		return m, wait(m.events)
	}

	return m, nil
}

func (m *Model) View() string {
	if m.latest != nil {
		return fmt.Sprintf("%s\n", m.latest.View())
	}

	return "no events yet"
}

// waits and blocks until a new event is received in order
// to return a bubbletea understandable tea.Cmd
func wait(c <-chan Event) tea.Cmd {
	return func() tea.Msg {
		return Event(<-c)
	}
}
