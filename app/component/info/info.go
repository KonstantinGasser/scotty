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

const (
	connected = iota
	paused
	disconnected

	symbolConnected    = "●"
	symbolPaused       = "◍"
	symbolDisconnected = "◌"
)

type stat struct {
	label string
	// colored string
	style     lipgloss.Style
	count     int
	state     int
	stateChar string
	color     lipgloss.Color
	compiled  string
}

func (s *stat) increment() *stat { s.count++; return s }
func (s *stat) compile() *stat {

	s.compiled = s.style.Render(fmt.Sprintf("%s %d", s.stateChar, s.count))
	return s
}

type Model struct {
	ready bool
	// baseInfo is the compiled string
	// for the active mode and the spacer
	// right next to it
	baseInfo string
	// each mode can have zero to many keystroke
	// options which are displayed in the info bar
	availOpts []string
	// statsMap keeps track of stats:index
	// for the stats slice allwoing for
	// O(1) search time
	statsMap map[string]int
	stats    []*stat
}

func New() *Model {
	return &Model{
		ready:    false,
		baseInfo: "",
		statsMap: make(map[string]int),
		stats:    []*stat{},
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
	case styles.Dimensions:
		if !model.ready {
			model.ready = true
		}

	case requestSubscribe:
		index, ok := model.statsMap[msg.label]
		if ok {
			model.stats[index].state = connected
			model.stats[index].stateChar = symbolConnected
			model.stats[index].compile()
			break
		}

		newStat := &stat{
			label:     msg.label,
			style:     lipgloss.NewStyle().Padding(0, 1).Foreground(msg.fg).Background(styles.BgFooter),
			state:     msg.state,
			stateChar: symbolConnected,
			count:     msg.count,
			color:     msg.fg,
			compiled:  "",
		}
		newStat.compile()

		model.stats = append(model.stats, newStat)
		model.statsMap[newStat.label] = len(model.stats) - 1

	case requestUnsubscribe:
		index, ok := model.statsMap[string(msg)]
		if !ok {
			break
		}
		model.stats[index].state = disconnected
		model.stats[index].stateChar = symbolDisconnected
		model.stats[index].compile()
	case requestPause:
		for index, st := range model.stats {
			if st.state == disconnected {
				continue
			}
			model.stats[index].state = paused
			model.stats[index].stateChar = symbolPaused
		}
	case requestResume:
		for index, st := range model.stats {
			if st.state == disconnected {
				continue
			}
			model.stats[index].state = connected
			model.stats[index].stateChar = symbolConnected
		}
	case requestIncrement:
		index, ok := model.statsMap[string(msg)]
		if !ok {
			break
		}
		model.stats[index].increment().compile()
	case requestMode:
		// a message received here effects the baseInfo of the model (the mode).
		// Options are stored seperatly and joined with the other information
		// on the call of View.

		model.baseInfo = lipgloss.NewStyle().
			Padding(0, 1).
			Bold(true).
			Foreground(lipgloss.Color("#ffffff")).
			Background(msg.bg).
			Render(msg.mode)

		if len(msg.opts) > 0 {
			model.availOpts = []string{}
			for _, opt := range msg.opts {
				model.availOpts = append(model.availOpts, lipgloss.NewStyle().Bold(true).Background(styles.BgFooter).Render(opt))
			}
		}
	}
	return model, tea.Batch(cmds...)
}

func (model Model) View() string {

	var (
		statsList, optsList string
	)

	statsTmp := []string{}
	for _, st := range model.stats {
		statsTmp = append(statsTmp, st.compiled)
	}

	statsList = lipgloss.JoinHorizontal(lipgloss.Left, statsTmp...)
	optsList = lipgloss.JoinHorizontal(lipgloss.Left, model.availOpts...)

	return lipgloss.JoinHorizontal(lipgloss.Left,
		model.baseInfo,
		statsList,
		optsList,
	)
}
