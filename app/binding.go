package app

import (
	"github.com/charmbracelet/bubbles/key"
)

type bindings struct {
	Up   key.Binding
	Down key.Binding
	Help key.Binding
	Quit key.Binding
}

func (b bindings) ShortHelp() []key.Binding {
	return []key.Binding{b.Help, b.Quit}
}

func (b bindings) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{b.Up, b.Down},
		{b.Help, b.Quit},
	}
}

var defaultBindings = bindings{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "move down"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "toggle help"),
	),
	Quit: key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("ctrl+c", "exit scotty"),
	),
}
