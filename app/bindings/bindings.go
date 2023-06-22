package bindings

import (
	"github.com/KonstantinGasser/scotty/debug"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kr/pretty"
)

type Func func(tea.KeyMsg) tea.Cmd

func (fn Func) Call(msg tea.KeyMsg) tea.Cmd {
	if fn == nil {
		return NilFunc.Call(msg)
	}
	return fn(msg)
}

var NilFunc Func = func(msg tea.KeyMsg) tea.Cmd {
	return nil
}

type Options map[string]*Node

type Node struct {
	binding      key.Binding
	bindingLabel string
	options      Options
	action       Func
}

func newNode(k string) *Node {
	return &Node{
		binding:      key.NewBinding(key.WithKeys(k)),
		bindingLabel: k,
		options:      Options{},
		action:       nil,
	}
}

func (node *Node) Option(k string) *Node {
	if opt, ok := node.options[k]; ok {
		return opt
	}

	optionNode := newNode(k)
	node.options[k] = optionNode

	return optionNode
}

func (node *Node) Action(act Func) *Node {
	node.action = act
	return node
}

type SequenceTree struct {
	onESC Func
	root  *Node
}

func newSeqTree(k string) SequenceTree {
	return SequenceTree{
		onESC: NilFunc,
		root:  newNode(k),
	}
}

func (seq *SequenceTree) OnESC(fn Func) *SequenceTree {
	seq.onESC = fn
	return seq
}

func (seq *SequenceTree) Option(k string) *Node {
	return seq.root.Option(k)
}

func (seq *SequenceTree) Action(act Func) *Node {
	seq.root.Action(act)
	return seq.root
}

type Binder interface {
	Bind(k string) *SequenceTree
}

type seqOption struct {
	onESC Func
	node  *Node
}
type Map struct {
	next  *seqOption
	binds map[string]*SequenceTree
}

func NewMap() *Map {
	return &Map{
		next:  nil,
		binds: map[string]*SequenceTree{},
	}
}

func (m *Map) Bind(k string) *SequenceTree {

	if seq, ok := m.binds[k]; ok {
		return seq
	}

	seq := newSeqTree(k)
	m.binds[k] = &seq

	return &seq
}

func (m *Map) Matches(msg tea.KeyMsg) bool {

	try := msg.String()

	// first we need to check if the past call to Matches
	// set options we and if the KeyMsg matches any of
	// these options
	if m.next != nil {
		// expend bindings by "esc"
		// which ends the sequence execution
		if try == "esc" {
			return true
		}

		next, ok := m.next.node.options[try]
		// if no option was found we do nothing.
		// user might have typed to wrong key or somthing
		// we keep the options as is.
		if !ok {
			return false
		}
		// user chose an option from the current next node
		m.next.node = next
		return true
	}

	// try and see if the key exists in the bindings map
	seq, ok := m.binds[try]
	if !ok {
		return false
	}

	// in case it does there might be further options
	// set on the root node
	if len(seq.root.options) > 0 {
		m.next = &seqOption{
			onESC: seq.onESC,
			node:  seq.root,
		}

	}

	return true
}

func (m *Map) Exec(msg tea.KeyMsg) Func {

	// we need to check if the KeyMsg matches the m.next.binding
	// if so return m.next.action. if not we need to check if the
	// KeyMsg machtes any m.next.options if so return m.next.options[x].action
	// and update the m.next with m.next.options[x]
	if m.next != nil {
		if key.Matches(msg, key.NewBinding(key.WithKeys("esc"))) {
			onESC := m.next.onESC
			m.next = nil
			return onESC
		}

		if key.Matches(msg, m.next.node.binding) {
			act := m.next.node.action
			if len(m.next.node.options) <= 0 {
				m.next = nil
			}
			return act
		}

		// next check and update options
		next, ok := m.next.node.options[msg.String()]
		if !ok {
			return NilFunc
		}

		if len(next.options) <= 0 {
			m.next = nil
			return next.action
		}

		m.next.node = next
		return m.next.node.action
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
