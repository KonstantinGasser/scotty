package bindings

import (
	"testing"

	// "github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	// "github.com/kr/pretty"
)

// func TestSimpleSeqTree(t *testing.T) {
//
// 	m := NewMap()
//
// 	m.Bind("f").Action(func(km tea.KeyMsg) tea.Cmd {
// 		return func() tea.Msg {
// 			return "hello world"
// 		}
// 	})
//
// 	pretty.Print(m.binds["f"])
// }

func TestOptionSeqTree(t *testing.T) {

	m := NewMap()

	// m.Bind(" ").Option("f").Action(func(km tea.KeyMsg) tea.Cmd {
	// 	return func() tea.Msg {
	// 		return "called: SPC->f"
	// 	}
	// })

	m.Bind(" ").Option("f").Action(func(km tea.KeyMsg) tea.Cmd {
		return func() tea.Msg {
			return "called: SPC"
		}
	})

	msgs := []tea.KeyMsg{
		{
			Type:  tea.KeySpace,
			Runes: []rune(" "),
			Alt:   false,
		},
		{
			Type:  tea.KeyRunes,
			Runes: []rune("f"),
			Alt:   false,
		},
	}

	// kSPC := key.NewBinding(key.WithKeys(" "))
	// kF := key.NewBinding(key.WithKeys("f"))
	// t.Logf("KeyMsg(SPC) == Binding(SPC) => %v\n", key.Matches(msgs[0], kSPC))
	// t.Logf("KeyMsg(f) == Binding(f) => %v\n", key.Matches(msgs[1], kF))

	for _, msg := range msgs {
		ok := m.Matches(msg)
		if !ok {
			t.Logf("Key: %q not found in Map\n", msg)
			continue
		}

		t.Logf("Key: %q found in Map. Action: %v\n", msg, m.Exec(msg).Call(msg))
	}
}
