package pager

import (
	tea "github.com/charmbracelet/bubbletea"
)

type command struct {
	width, height int
}

func newCommand(w, h int) *command {
	return &command{
		width:  w,
		height: h,
	}
}

func (cmd *command) Init() tea.Cmd {
	return nil
}

func (cmd *command) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	var (
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		cmd.width = msg.Width - 2 // account for margin
		cmd.height = msg.Height
		return cmd, nil
	}

	return cmd, tea.Batch(cmds...)
}

func (cmd *command) View() string {
	return "command input something"
}
