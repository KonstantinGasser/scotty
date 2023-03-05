package event

import tea "github.com/charmbracelet/bubbletea"

type ReloadBuffer struct{}

func RequestReload() tea.Cmd {
	return func() tea.Msg {
		return ReloadBuffer{}
	}
}
