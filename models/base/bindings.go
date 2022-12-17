package base

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type keyMap struct {
	Up   key.Binding
	Down key.Binding
	View key.Binding
	Help key.Binding
	Quit key.Binding
}

// ShortHelp returns keybindings to be shown in the mini help view. It's part
// of the key.Map interface.
func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit}
}

// FullHelp returns keybindings for the expanded help view. It's part of the
// key.Map interface.
func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.View}, // first column
		{k.Help, k.Quit},       // second column
	}
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "move down"),
	),
	View: key.NewBinding(
		key.WithKeys("v"),
		key.WithHelp("v", "view log"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "toggle help"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "esc", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
}

func (m *Model) resolveBinding(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch true {
	case key.Matches(msg, m.bindings.Quit):
		m.quite <- struct{}{}
		return m, tea.Quit
	case key.Matches(msg, m.bindings.Help):
		m.help.ShowAll = !m.help.ShowAll
	case key.Matches(msg, m.bindings.Down):
		if m.index <= 0 {
			break
		}
		m.index++
	case key.Matches(msg, m.bindings.Up):
		if m.index >= len(m.logs) {
			break
		}
		m.index--
	case key.Matches(msg, m.bindings.View):

	}

	return m, nil
}
