package pager

import (
	"fmt"
	"strconv"

	"github.com/KonstantinGasser/scotty/app/styles"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wrap"
)

var (
	commandStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(styles.DefaultColor.Border)

	parsedStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(styles.DefaultColor.Border)

	emptyParsedMsg = lipgloss.NewStyle().
			Bold(true).
			Foreground(styles.DefaultColor.Light).
			Render("no value which can be formatted")
)

// parsedIndex holds the parsed log data
type parsedLog struct {
	index int
	label string
	data  []byte
}
type parserIndex int

// send whenever an input is provided and the
// confirmed by the enter key
func emitIndex(index int) tea.Cmd {
	return func() tea.Msg {
		return parserIndex(index)
	}
}

func emitParsed(v *parsedLog) tea.Cmd {
	return func() tea.Msg {
		return v
	}
}

type command struct {
	width, height int
	input         textinput.Model

	parsedItem *parsedLog
	parsedView viewport.Model
	err        error
}

func newCommand(w, h int) *command {
	width := int(w/3) - 1
	input := textinput.New()
	input.Placeholder = "line number (use k/j to move and ESC to exit)"
	input.Prompt = ":"
	return &command{
		width:      width,
		height:     h,
		input:      input,
		parsedItem: nil,
		parsedView: defaultLogView(width, 1),
		err:        nil,
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
	case tea.MouseMsg:
		c.parsedView, cmd = c.parsedView.Update(msg)
		cmds = append(cmds, cmd)

	case tea.WindowSizeMsg:
		c.width = int(msg.Width/3) - 1
		c.height = msg.Height - bottomSectionHeight - magicNumber
		return c, nil
	case tea.KeyMsg:
		switch msg.String() {
		// same as for the char ":" we don't want to propagate the keystrokes
		// down to the textinput.Model since these keystrokes are registered as
		// normal user input.
		case "j", "k":
			c.input.Reset()
			return c, tea.Batch(cmds...)

		case "esc":
			c.input.Blur()
			c.parsedView.SetContent("")
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
			cmds = append(cmds, emitIndex(index))
		}
	case *parsedLog:
		c.parsedItem = msg
	}

	c.input, cmd = c.input.Update(msg)
	cmds = append(cmds, cmd)

	return c, tea.Batch(cmds...)
}

func (c *command) View() string {
	if c.err != nil {
		return commandStyle.
			Width(c.width).
			Background(styles.DefaultColor.Error).
			Render(c.err.Error())
	}

	c.parsedView.SetContent(emptyParsedMsg)
	parsedContent := c.parsedView.View()
	if c.parsedItem != nil {
		value := lipgloss.NewStyle().
			MarginTop(1).
			Render(
				wrap.String(string(c.parsedItem.data), c.width-2),
			)

		c.parsedView.Height = lipgloss.Height(value)
		c.parsedView.SetContent(value)

		parsedContent = lipgloss.JoinVertical(lipgloss.Left,
			"["+strconv.Itoa(c.parsedItem.index)+"]"+c.parsedItem.label,
			c.parsedView.View(),
		)
	}

	return lipgloss.JoinVertical(lipgloss.Top,
		commandStyle.
			Width(c.width).
			Render(
				c.input.View(),
			),
		parsedStyle.
			Padding(0, 1).
			Width(c.width).
			Render(
				parsedContent,
			),
	)
}

func defaultLogView(width, height int) viewport.Model {
	vp := viewport.New(width, height)
	vp.MouseWheelEnabled = true

	return vp
}
