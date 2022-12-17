package ui

import (
	"github.com/KonstantinGasser/scotty/streams"
	"github.com/KonstantinGasser/scotty/ui/common"
	"github.com/KonstantinGasser/scotty/ui/components/footer"
	"github.com/KonstantinGasser/scotty/ui/components/header"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	margin = 2
)

type UI struct {
	quite chan<- struct{}

	// general settings
	width, height int
	bindings      common.Bindings
	help          help.Model

	// components
	header *header.Model
	footer footer.Model

	// async actions
	messages <-chan streams.Message
}

func New(w, h int, q chan<- struct{}, msgs <-chan streams.Message) *UI {
	w, h = w-margin, h-margin

	vp := viewport.New(w, h)
	vp.MouseWheelEnabled = true
	vp.Style = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("45"))
	vp.SetContent("beam me up, Scotty!...")
	vp.HighPerformanceRendering = true

	termUI := UI{
		quite: q,

		// general settings
		width:    w,
		height:   h,
		bindings: common.DefaultBindings,
		help:     help.New(),

		// components
		header: header.New(w, h, "hello world"),
		footer: footer.New(w, h),

		// async actions
		messages: msgs,
	}

	return &termUI
}

func (u UI) Init() tea.Cmd {
	return u.consume
}

func (u *UI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {
	case tea.KeyMsg:
		return u.resolveBinding(msg)
	case tea.WindowSizeMsg:
		u.width, u.height = msg.Width-margin, msg.Height-margin

		u.header.SetSize(u.width, u.height)
	case streams.Message:
		return u, u.consume
	}

	return u, nil
}

var consistentPadding = lipgloss.NewStyle() //.PaddingLeft(2).PaddingRight(2)

func (u UI) View() string {

	viewAndHeader := lipgloss.JoinVertical(lipgloss.Left, u.header.View(), "nothing there lul")

	return lipgloss.JoinVertical(lipgloss.Left,
		consistentPadding.Render(viewAndHeader),
		consistentPadding.Render(u.footer.View()),
	)
}

func (u *UI) resolveBinding(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch true {
	case key.Matches(msg, u.bindings.Quit):
		u.quite <- struct{}{}
		return u, tea.Quit
	case key.Matches(msg, u.bindings.Help):
		u.help.ShowAll = !u.help.ShowAll
	}

	return u, nil
}

func (u *UI) consume() tea.Msg {
	return <-u.messages
}
