package app

import (
	"github.com/charmbracelet/bubbles/key"
)

type bindings struct {
	Up    key.Binding
	Down  key.Binding
	Input key.Binding
	Exit  key.Binding
	Quit  key.Binding
}

func (b bindings) ShortHelp() []key.Binding {
	return []key.Binding{b.Quit}
}

func (b bindings) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{b.Up, b.Down},
		{b.Quit},
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
	Input: key.NewBinding(
		key.WithKeys(":"),
	),
	Exit: key.NewBinding(
		key.WithKeys("esc"),
		key.WithKeys("q"),
	),
	Quit: key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("ctrl+c", "exit scotty"),
	),
}
