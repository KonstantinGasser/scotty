package event

import (
	tea "github.com/charmbracelet/bubbletea"
)

type GlobalRefreshBuffer struct{}

func RequestGlobalBufferRefresh() tea.Msg {
	return GlobalRefreshBuffer{}
}
