package event

import (
	tea "github.com/charmbracelet/bubbletea"
)

type BlockKeys []string

func BlockKeysRequest(keys ...string) tea.Cmd {
	return func() tea.Msg {
		return BlockKeys(keys)
	}
}

type Increment string

func IncrementRequest(label string) tea.Cmd {
	return func() tea.Msg {
		return Increment(label)
	}
}
