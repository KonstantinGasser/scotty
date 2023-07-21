package info

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type requestSubscribe struct {
	label string
	state int
	count int
	fg    lipgloss.Color
}

func RequestSubscribe(label string, fg lipgloss.Color) tea.Cmd {
	return func() tea.Msg {
		return requestSubscribe{
			label: label,
			state: connected,
			count: 0,
			fg:    fg,
		}
	}
}

type requestUnsubscribe string

func RequestUnsubscribe(label string) tea.Cmd {
	return func() tea.Msg {
		return requestUnsubscribe(label)
	}
}

type requestIncrement string

func RequestIncrement(label string) tea.Cmd {
	return func() tea.Msg {
		return requestIncrement(label)
	}
}

type requestPause struct{}

func RequestPause() tea.Cmd {
	return func() tea.Msg {
		return requestPause{}
	}
}

type requestResume struct{}

func RequestResume() tea.Cmd {
	return func() tea.Msg {
		return requestResume{}
	}
}

type AppMode struct {
	Label string
	Bg    lipgloss.Color
	Opts  []string
}

var (
	ModeFollowing    AppMode = AppMode{Label: "FOLLOWING", Bg: lipgloss.Color("#98c379"), Opts: []string{" ·p pause/continue", " ·g go to latest"}}
	ModeBrowsing     AppMode = AppMode{Label: "BROWSING", Bg: lipgloss.Color("#98c378"), Opts: []string{" ·j next", " ·k previous", " ·r reload"}}
	ModePaused       AppMode = AppMode{Label: "PAUSED", Bg: lipgloss.Color("#ff9640")}
	ModeGlobalCmd    AppMode = AppMode{Label: "GLOBAL", Bg: lipgloss.Color("54"), Opts: []string{" ·f follow", "·b browse", "·besc exit mode"}}
	ModePromptActive AppMode = AppMode{Label: "INPUT (exit with ESC)", Bg: lipgloss.Color("54"), Opts: []string{"·besc exit input mode"}}
)

type requestMode struct {
	mode string
	bg   lipgloss.Color
	opts []string
}

func RequestMode(mode AppMode) tea.Cmd {
	return func() tea.Msg {
		return requestMode{
			mode: strings.ToUpper(mode.Label),
			bg:   mode.Bg,
			opts: mode.Opts,
		}
	}
}
