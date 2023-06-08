package info

import (
	"fmt"

	"github.com/KonstantinGasser/scotty/app/styles"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	minHeight = 15
)

var (
	style = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder())
)

const (
	connected = iota
	paused
	disconnected

	symbolConnected    = "●"
	symbolPaused       = "◍"
	symbolDisconnected = "◌"
)

type stat struct {
	label   string
	colored string
	style   lipgloss.Style
	count   int
	state   int
}

func (s *stat) increment() { s.count++ }

type Model struct {
	ready bool
	// width, height int
	stats   map[string]*stat
	ordered []string
	mode    string
}

func New() *Model {
	return &Model{
		ready: false,
		// width:   0,
		// height:  0,
		stats:   map[string]*stat{},
		ordered: []string{},
		mode:    "",
	}
}

func (model Model) Init() tea.Cmd {
	return nil
}

func (model *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	var (
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case styles.FooterLine:
		if !model.ready {
			model.ready = true
		}

		// model.width = msg.Width()
		// model.height = msg.Height()
	case requestSubscribe:
		if _, ok := model.stats[msg.label]; ok {
			model.stats[msg.label].state = connected
			break
		}
		model.stats[msg.label] = &stat{
			label: msg.label,
			style: lipgloss.NewStyle().Padding(0, 1).Foreground(msg.fg).Background(styles.BgFooter),
			state: msg.state,
			count: msg.count,
		}
		model.ordered = append(model.ordered, msg.label)
	case requestUnsubscribe:
		if _, ok := model.stats[string(msg)]; !ok {
			break
		}
		model.stats[string(msg)].state = disconnected
	case requestPause:
		for label := range model.stats {
			if model.stats[label].state == disconnected {
				continue
			}
			model.stats[label].state = paused
		}
	case requestResume:
		for label := range model.stats {
			if model.stats[label].state == disconnected {
				continue
			}
			model.stats[label].state = connected
		}
	case requestIncrement:
		if beam, ok := model.stats[string(msg)]; ok {
			beam.increment()
		}
	case requestMode:
		model.mode = lipgloss.NewStyle().
			Padding(0, 1).
			Bold(true).
			Foreground(lipgloss.Color("#ffffff")).
			Background(msg.bg).
			Render(msg.mode)
	}
	return model, tea.Batch(cmds...)
}

func (model Model) View() string {

	var items = make([]string, len(model.ordered)+2) // +1 for the mode at the beginning of the line +1 for the spacer between mode and stats
	items[0] = model.mode
	items[1] = styles.SingleLineSpacer(5).Background(styles.BgFooter).Render("")
	var status = symbolConnected
	for i, label := range model.ordered {

		stat := model.stats[label]

		switch model.stats[label].state {
		case disconnected:
			status = symbolDisconnected
		case paused:
			status = symbolPaused
		}

		items[i+2 /*+2 as the zero/one items is the mode and spacer*/] = stat.style.Render(fmt.Sprintf("%s %d", status, stat.count))
	}

	list := lipgloss.JoinHorizontal(lipgloss.Top, items...)

	return list
}

func max(upper int, compare int) int {
	if compare > upper {
		return upper
	}
	return compare
}

func min(lower int, compare int) int {
	if compare < lower {
		return lower
	}
	return compare
}
