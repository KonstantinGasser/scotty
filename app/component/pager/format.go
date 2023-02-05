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
	formatModelStyle = lipgloss.NewStyle().
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

type formatter struct {
	width, height int
	input         textinput.Model

	parsedItem *parsedLog
	parsedView viewport.Model
	err        error
}

func newFormatter(w, h int) *formatter {
	width := int(w/3) - 1
	input := textinput.New()
	input.Placeholder = "line number (use k/j to move and ESC to exit)"
	input.Prompt = ":"
	return &formatter{
		width:      width,
		height:     h,
		input:      input,
		parsedItem: nil,
		parsedView: defaultLogView(width, 1),
		err:        nil,
	}
}

func (f *formatter) Init() tea.Cmd {
	return nil
}

func (f *formatter) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	var (
		cmds []tea.Cmd
		cmd  tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.MouseMsg:
		f.parsedView, cmd = f.parsedView.Update(msg)
		cmds = append(cmds, cmd)

	case tea.WindowSizeMsg:
		f.width = int(msg.Width/3) - 1
		f.parsedView.Width = f.width

		f.height = msg.Height - bottomSectionHeight - magicNumber
		return f, nil
	case tea.KeyMsg:
		switch msg.String() {
		// same as for the char ":" we don't want to propagate the keystrokes
		// down to the textinput.Model since these keystrokes are registered as
		// normal user input.
		case "j", "k":
			f.input.Reset()
			return f, tea.Batch(cmds...)

		case "esc":
			f.input.Blur()
			f.parsedView.SetContent("")
			return f, nil
		case ":":
			f.input.Reset()

			if f.input.Focused() {
				f.err = nil
				f.input.Blur()
				break
			}

			f.input.Focus()
			cmds = append(cmds, textinput.Blink)
			return f, tea.Batch(cmds...) // we want to ignore the update of the textinput.Model else ":" is registered as first key stroke of the input
		case "enter":
			value := f.input.Value()
			index, err := strconv.Atoi(value)
			if err != nil {
				f.err = fmt.Errorf("input %q is not numerif. Type the index of the line you want to parse", value)
				break
			}
			cmds = append(cmds, emitIndex(index))
		}
	case *parsedLog:
		f.parsedItem = msg
	}

	f.input, cmd = f.input.Update(msg)
	cmds = append(cmds, cmd)

	return f, tea.Batch(cmds...)
}

func (f *formatter) View() string {
	if f.err != nil {
		return formatModelStyle.
			Width(f.width).
			Background(styles.DefaultColor.Error).
			Render(f.err.Error())
	}

	f.parsedView.SetContent(emptyParsedMsg)
	parsedContent := f.parsedView.View()
	if f.parsedItem != nil {
		value := lipgloss.NewStyle().
			MarginTop(1).
			Render(
				wrap.String(string(f.parsedItem.data), f.width),
			)

		f.parsedView.Height = lipgloss.Height(value)
		f.parsedView.SetContent(value)

		parsedContent = lipgloss.JoinVertical(lipgloss.Left,
			"["+strconv.Itoa(f.parsedItem.index)+"]"+f.parsedItem.label,
			f.parsedView.View(),
		)
	}

	return lipgloss.JoinVertical(lipgloss.Top,
		formatModelStyle.
			Width(f.width).
			Render(
				f.input.View(),
			),
		parsedStyle.
			Padding(0, 1).
			Width(f.width).
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
