package tailing

import (
	tea "github.com/charmbracelet/bubbletea"
)

type forceRefresh struct{}

func ForceRefresh() tea.Cmd {
	return func() tea.Msg {
		return forceRefresh{}
	}
}
