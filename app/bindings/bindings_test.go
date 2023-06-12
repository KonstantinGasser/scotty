package bindings

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kr/pretty"
)

func TestSimpleSeqTree(t *testing.T) {

	m := NewMap()

	m.Bind("f").Action(func(km tea.KeyMsg) tea.Cmd {
		return func() tea.Msg {
			return "hello world"
		}
	})

	pretty.Print(m.binds["f"])
}

func TestOptionSeqTree(t *testing.T) {

	m := NewMap()

	m.Bind("SPC").Option("f").Action(func(km tea.KeyMsg) tea.Cmd {
		return func() tea.Msg {
			return "called: SPC->f"
		}
	})
	m.Bind("SPC").Option("g").Action(func(km tea.KeyMsg) tea.Cmd {
		return func() tea.Msg {
			return "called: SPC->g"
		}
	})

	pretty.Print(m.binds["SPC"])
}
