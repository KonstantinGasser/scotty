package pager

import (
	"fmt"
	"strconv"

	"github.com/KonstantinGasser/scotty/app/styles"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type parserIndex int

// send whenever an input is provided and the
// confirmed by the enter key
func parseLog(index int) tea.Cmd {
	return func() tea.Msg {
		return parserIndex(index)
	}
}

type command struct {
	width, height int
	input         textinput.Model
	err           error
}

func newCommand(w, h int) *command {
	input := textinput.New()
	input.Placeholder = "type the line number to parse as JSON"
	input.Prompt = ":"

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
			c.input.Reset()

			if c.input.Focused() {
				c.err = nil
				c.input.Blur()
				break
			}

			c.input.Focus()
			cmds = append(cmds, textinput.Blink)
			return c, tea.Batch(cmds...) // we want to ignore the update of the textinput.Model else ":" is registered as first key stroke of the input
		case "enter":
			value := c.input.Value()
			index, err := strconv.Atoi(value)
			if err != nil {
				c.err = fmt.Errorf("input %q is not numeric. Type the index of the line you want to parse", value)
				break
			}
			cmds = append(cmds, parseLog(index))
		}
	}

	c.input, cmd = c.input.Update(msg)
	cmds = append(cmds, cmd)

	return c, tea.Batch(cmds...)
}

func (c *command) View() string {
	if c.err != nil {
		return lipgloss.NewStyle().
			Foreground(styles.ColorError).
			Render(c.err.Error())
	}

	return c.input.View()
}
