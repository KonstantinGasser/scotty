package ui

import (
	"fmt"

	"github.com/KonstantinGasser/scotty/ui/common"
	"github.com/KonstantinGasser/scotty/ui/components/header"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	margin = 2
)

type UI struct {
	quite chan<- struct{}

	bindings common.Bindings
	help     help.Model

	header tea.Model
	text   string
}

func New(w, h int, q chan<- struct{}) *UI {
	w, h = w-margin, h-margin

	termUI := UI{
		quite: q,

		bindings: common.DefaultBindings,
		help:     help.New(),

		header: header.New(w, h, "hello world"),

		text: fmt.Sprintf("Width: %d, Height: %d", w, h),
	}

	return &termUI
}

func (u UI) Init() tea.Cmd {
	return nil
}

func (u *UI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {
	case tea.KeyMsg:
		return u.resolveBinding(msg)
	}
	return u, nil
}

func (u UI) View() string {

	view := u.text

	viewAndHeader := lipgloss.JoinVertical(lipgloss.Left, u.header.View(), view)
	return lipgloss.JoinVertical(lipgloss.Left, viewAndHeader, u.help.View(u.bindings))
}

func (u *UI) resolveBinding(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch true {
	case key.Matches(msg, u.bindings.Quit):
		u.quite <- struct{}{}
		return u, tea.Quit
	case key.Matches(msg, u.bindings.Help):
		u.help.ShowAll = !u.help.ShowAll
	}

	u.text = "pressed: " + msg.String()
	return u, nil
}
