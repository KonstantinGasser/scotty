package event

import tea "github.com/charmbracelet/bubbletea"

type ReloadBuffer struct{}

func RequestReload() tea.Cmd {
	return func() tea.Msg {
		return ReloadBuffer{}
	}
}

type FormatInit int

func RequestFormatInit(index int) tea.Cmd {
	return func() tea.Msg {
		return FormatInit(index)
	}
}

type FormatNext struct{}

func RequestFormatNext() tea.Cmd {
	return func() tea.Msg {
		return FormatNext{}
	}
}

type FormatPrevious struct{}

func RequestFormatPrevious() tea.Cmd {
	return func() tea.Msg {
		return FormatPrevious{}
	}
}

type DimensionMsg struct {
	AvailableWidth  int
	AvailableHeight int
}
