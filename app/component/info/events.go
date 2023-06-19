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

type requestMode struct {
	mode string
	bg   lipgloss.Color
	opts []string
}

func RequestMode(mode string, bg lipgloss.Color, opts ...string) tea.Cmd {
	return func() tea.Msg {
		return requestMode{
			mode: strings.ToUpper(mode),
			bg:   bg,
			opts: opts,
		}
	}
}
