package pager

import (
	"github.com/KonstantinGasser/scotty/app/styles"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	marginLeft   = 2
	marginRight  = 2
	marginTop    = 1
	marginBottom = 1
)

var (
	pagerStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(
			styles.ColorBorder,
		)

	pagerStyleActive = pagerStyle.Copy().
				UnsetBorderForeground().
				BorderForeground(
			styles.ColorBorderActive,
		)
)

type Logger struct {

	// underlying model which handles
	// scrolling and rendering of the logs
	vp viewport.Model

	// available tty width and height
	// updates if changes
	width, height int
}

func NewLogger(width, height int) *Logger {

	w, h := width-(marginLeft+marginRight), height-(marginTop+marginBottom)
	vp := viewport.New(w, h)

	return &Logger{
		vp:     vp,
		width:  w,
		height: h,
	}
}

func (log *Logger) Init() tea.Cmd {
	return nil
}

func (log *Logger) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	return log, nil
}

func (log *Logger) View() string {
	return pagerStyle.Render(
		log.vp.View(),
	)
}
