package browsing

import "github.com/charmbracelet/bubbles/key"

type bindings struct {
	Focus key.Binding
	Enter key.Binding
}

var defaultBindings = bindings{
	Focus: key.NewBinding(
		key.WithKeys(":"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
	),
}
