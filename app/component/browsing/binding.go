package browsing

import "github.com/charmbracelet/bubbles/key"

type bindings struct {
	Focus key.Binding
	Enter key.Binding
	Up    key.Binding
	Down  key.Binding
	Exit  key.Binding
}

var defaultBindings = bindings{
	Focus: key.NewBinding(
		key.WithKeys(":"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
	),
	Up: key.NewBinding(
		key.WithKeys("k"),
	),
	Down: key.NewBinding(
		key.WithKeys("j"),
	),
	Exit: key.NewBinding(
		key.WithKeys("esc"),
	),
}
