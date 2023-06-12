package bindings

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type Func func(tea.KeyMsg) tea.Cmd

var NilFunc Func = func(tea.KeyMsg) tea.Cmd { return nil }

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
