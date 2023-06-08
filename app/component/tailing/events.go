package tailing

import (
	tea "github.com/charmbracelet/bubbletea"
)

type forceRefresh struct{}

func RequestRefresh() tea.Cmd {
	return func() tea.Msg {
		return forceRefresh{}
	}
}

type PauseRequest struct{}

func RequestPause() tea.Cmd {
	return func() tea.Msg {
		return PauseRequest{}
	}
}

type ResumeRequest struct{}

func RequestResume() tea.Cmd {
	return func() tea.Msg {
		return ResumeRequest{}
	}
}
