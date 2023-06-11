package bindings

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/exp/slices"
)

type Mapper interface {
	Map(k key.Binding, fn Func)
	Disable(ks ...key.Binding)
	Enable(ks ...key.Binding)
	Cmd(msg tea.KeyMsg) Func
}

const (
	keySPC = " "

	// if set implies that keySPC has be clicked
	// which can change the actions of certain local
	// bindings
	stateGlobal = iota + 1
	stateLocal
)

var global = key.NewBinding(
	key.WithKeys(keySPC),
	key.WithHelp("SPC", "start of global command"),
)

type Func func(tea.KeyMsg) tea.Cmd

func (fn Func) Call(msg tea.KeyMsg) tea.Cmd { return fn(msg) }

type KeyMap struct {
	state uint8
	binds map[*key.Binding]Func
}

func New() *KeyMap {
	quite := key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("ctrl+c", "exit scotty"),
	)

	keys := KeyMap{
		state: stateLocal,
		binds: map[*key.Binding]Func{
			&quite: func(tea.KeyMsg) tea.Cmd { return tea.Quit },
		},
	}

	return &keys
}

func (keys *KeyMap) Map(k key.Binding, fn Func) {
	keys.binds[&k] = fn
}

func (keys *KeyMap) Match(k tea.KeyMsg) bool {
	str := k.String()

	for binding := range keys.binds {
		for _, key := range binding.Keys() {
			if key == str {
				return true
			}
		}
	}
	return false
}

// TODO @KonstantinGasser:
// wow, there must be a better way to do things???
// maybey use a different data structre?
func (keys *KeyMap) Disable(ks ...key.Binding) {
	for _, k := range ks {
		for binding := range keys.binds {
			findAndFlipOnOff(binding, k.Keys(), false)
		}
	}
}

func (keys *KeyMap) Enable(ks ...key.Binding) {
	for _, k := range ks {
		for binding := range keys.binds {
			findAndFlipOnOff(binding, k.Keys(), true)
		}
	}
}

func findAndFlipOnOff(b *key.Binding, keys []string, onOff bool) {
	for _, kVal := range keys {
		if slices.Contains(b.Keys(), kVal) {
			b.SetEnabled(onOff)
		}
	}

}

func (keys *KeyMap) Cmd(msg tea.KeyMsg) Func {
	for binding, cmd := range keys.binds {
		for _, key := range binding.Keys() {
			if key == msg.String() {
				return cmd
			}
		}
	}
	return func(tea.KeyMsg) tea.Cmd { return nil }
}
