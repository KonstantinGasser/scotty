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

	parsedStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(styles.ColorBorder)
)

// parsedIndex holds the parsed log data
type parsedLog []byte
type parserIndex int

// send whenever an input is provided and the
// confirmed by the enter key
func emitIndex(index int) tea.Cmd {
	return func() tea.Msg {
		return parserIndex(index)
	}
}

func emitParsed(v []byte) tea.Cmd {
	return func() tea.Msg {
		return parsedLog(v)
	}
}

type command struct {
	width, height int
	input         textinput.Model
	parsed        []byte
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

	debug.Print("[cmd.Update] %T - %v\n", msg, msg)
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
			c.parsed = nil
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
			debug.Print("[enter] index=%d\n", index)
			cmds = append(cmds, emitIndex(index))
		}
	case parsedLog:
		debug.Print("[parsedLog] %s\n", msg)
		c.parsed = []byte(msg)
	}

	c.input, cmd = c.input.Update(msg)
	cmds = append(cmds, cmd)

	return c, tea.Batch(cmds...)
}

func (c *command) View() string {
	if c.err != nil {
		return commandStyle.
			Width(int(c.width/3) - 1).
			Background(styles.ColorError).
			Render(c.err.Error())
	}

	// debug.Print("cmd with width: %d\n", int(c.width/3)-1)
	debug.Print("parsed: %s\n", c.parsed)

	if c.parsed != nil {
		return lipgloss.JoinVertical(lipgloss.Top,
			commandStyle.
				Width(int(c.width/3)-1).
				Render(
					c.input.View(),
				),
			parsedStyle.
				Width(int(c.width/3)-1).
				Height(lipgloss.Height(string(c.parsed))).
				Padding(1).
				Render(
					string(c.parsed),
				),
		)
	}
	return commandStyle.
		Width(int(c.width/3) - 1).
		Render(
			c.input.View(),
		)
}
