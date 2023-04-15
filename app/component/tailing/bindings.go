package tailing

import "github.com/charmbracelet/bubbles/key"

type bindings struct {
	Pause       key.Binding
	FastForward key.Binding
}

var defaultBindings = bindings{
	Pause: key.NewBinding(
		key.WithKeys("p"),
	),
	FastForward: key.NewBinding(
		key.WithKeys("g"),
	),
}
