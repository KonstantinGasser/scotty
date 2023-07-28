package bindings

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestOptionSeqTree(t *testing.T) {

	m := NewMap()

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

	for _, msg := range msgs {
		ok := m.Matches(msg)
		if !ok {
			t.Logf("Key: %q not found in Map\n", msg)
			continue
		}

		t.Logf("Key: %q found in Map. Action: %v\n", msg, m.Exec(msg).Call(msg))
	}
}

func TestOnESCForSequence(t *testing.T) {

	m := NewMap()

	m.Bind(" ").OnESC(func(msg tea.KeyMsg) tea.Cmd {
		return func() tea.Msg { return 1 }
	}).Option("f").Action(func(km tea.KeyMsg) tea.Cmd { return nil })

	msgs := []tea.KeyMsg{
		{
			Type:  tea.KeySpace,
			Runes: []rune(" "),
			Alt:   false,
		},
		{
			Type:  tea.KeyEsc,
			Runes: []rune("esc"),
			Alt:   false,
		},
	}

	for _, msg := range msgs {
		if ok := m.Matches(msg); !ok && msg.String() != "esc" {
			t.Fatalf("Key: %q invalid key for sequence", msg)
		}

		call := m.Exec(msg)
		if msg.String() == "esc" {
			resp := call.Call(msg)
			if resp == nil {
				t.Fatalf("onESC CMD is nil!")
			}
			if val := resp(); val != 1 {
				t.Fatalf("ESC call returned not expected value: got: %v, wanted: %d", val, 1)
			}
		}
	}

}
