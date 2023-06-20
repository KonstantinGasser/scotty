package browsing

import tea "github.com/charmbracelet/bubbletea"

type initView struct{}

func RequestInitialView() tea.Msg {
	return initView{}
}
