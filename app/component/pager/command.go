package pager

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type command struct {
	width, height int
	input         textinput.Model
	err           error
}

func newCommand(w, h int) *command {
	input := textinput.New()
	input.Placeholder = "type the line number to parse as JSON"

	return &command{
		width:  w,
		height: h,
		input:  input,
		err:    nil,
	}
}

func (cmd *command) Init() tea.Cmd {
	return nil
}

func (c *command) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	var (
		cmds []tea.Cmd
		cmd  tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		c.width = msg.Width - 2 // account for margin
		c.height = msg.Height
		return c, nil
	case tea.KeyMsg:
		switch msg.String() {
		case ":":
			if c.input.Focused() {
				c.input.Reset()
				break
			}

			c.input.Focus()
			cmds = append(cmds, textinput.Blink)
		}
	}

	c.input, cmd = c.input.Update(msg)
	cmds = append(cmds, cmd)

	return c, tea.Batch(cmds...)
}

func (c *command) View() string {
	if c.err != nil {
		c.input.Err = c.err
	}

	return c.input.View()
}
