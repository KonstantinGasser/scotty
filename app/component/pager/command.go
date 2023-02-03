package pager

import (
	"fmt"
	"strconv"

	"github.com/KonstantinGasser/scotty/app/styles"
	"github.com/KonstantinGasser/scotty/debug"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	commandStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.ColorBorder)
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
		c.width = msg.Width
		c.height = msg.Height
		return c, nil
	case tea.KeyMsg:
		switch msg.String() {
		// same as for the char ":" we don't want to propagate the keystrokes
		// down to the textinput.Model since these keystrokes are registered as
		// normal user input.
		case "j", "k":
			return c, tea.Batch(cmds...)

		case "esc":
			c.input.Blur()
			return c, nil
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
		return commandStyle.
			Width(int(c.width / 3)).
			Background(styles.ColorError).
			Render(c.err.Error())
	}

	debug.Print("cmd with width: %d\n", int(c.width/3)-1)
	return commandStyle.
		Width(int(c.width/3) - 1).
		Render(
			c.input.View(),
		)
}
