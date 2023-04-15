package app

import (
	"github.com/charmbracelet/bubbles/key"
)

type bindings struct {
	Quit      key.Binding
	SwitchTab key.Binding
}

func (b bindings) ShortHelp() []key.Binding {
	return []key.Binding{b.Quit}
}

var defaultBindings = bindings{
	Quit: key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("ctrl+c", "exit scotty"),
	),
	SwitchTab: key.NewBinding(
		key.WithKeys("1", "2", "3"),
	),
}
