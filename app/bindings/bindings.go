package bindings

import (
	"github.com/KonstantinGasser/scotty/debug"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kr/pretty"
)

type Func func(tea.KeyMsg) tea.Cmd

func (fn Func) Call(msg tea.KeyMsg) tea.Cmd {
	return fn(msg)
}

var NilFunc Func = func(msg tea.KeyMsg) tea.Cmd {
	debug.Print("call to NilFunc for Key %q\n", msg.String())
	return nil
}

type Options map[string]*Node

type Node struct {
	binding key.Binding
	options Options
	action  Func
}

func newNode(k string) *Node {
	return &Node{
		binding: key.NewBinding(key.WithKeys(k)),
		options: Options{},
		action:  NilFunc,
	}
}

func (node *Node) Option(k string) *Node {
	if opt, ok := node.options[k]; ok {
		return opt
	}

	node.options[k] = newNode(k)
	return node
}

func (node *Node) Action(act Func) *Node {
	node.action = act
	return node
}

type SequenceTree struct {
	root *Node
}

func newSeqTree(k string) SequenceTree {
	return SequenceTree{
		root: newNode(k),
	}
}

func (seq *SequenceTree) Option(k string) *Node {
	seq.root.Option(k)
	return seq.root
}

func (seq *SequenceTree) Action(act Func) *Node {
	seq.root.Action(act)
	return seq.root
}

type Binder interface {
	Bind(k string) *SequenceTree
}

type Map struct {
	activeOpts *Node
	binds      map[string]SequenceTree
}

func NewMap() *Map {
	return &Map{
		activeOpts: nil,
		binds:      map[string]SequenceTree{},
	}
}

func (m *Map) Bind(k string) *SequenceTree {
	if seq, ok := m.binds[k]; ok {
		return &seq
	}

	seq := newSeqTree(k)
	m.binds[k] = seq

	return &seq
}

func (m *Map) Matches(msg tea.KeyMsg) bool {
	if m.activeOpts != nil {
		_, ok := m.activeOpts.options[msg.String()]
		return ok
	}

	seq, ok := m.binds[msg.String()]
	if !ok {
		return false
	}
	m.activeOpts = seq.root.options[msg.String()]
	return true
}

func (m *Map) Exec(msg tea.KeyMsg) Func {

	if m.activeOpts != nil {
		if len(m.activeOpts.options) == 0 {
			return m.activeOpts.action
		}

		opt, ok := m.activeOpts.options[msg.String()]
		if !ok {
			return NilFunc
		}

		m.activeOpts = opt
		return opt.action
	}

	seq, ok := m.binds[msg.String()]
	if !ok {
		return NilFunc
	}

	return seq.root.action
}

func (m *Map) Debug() {
	debug.Print("AST of bindings.Map:\n%s\n", pretty.Sprint(*m))
}
