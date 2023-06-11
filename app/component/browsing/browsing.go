package browsing

import (
	"strconv"

	"github.com/KonstantinGasser/scotty/app/bindings"
	"github.com/KonstantinGasser/scotty/app/styles"
	"github.com/KonstantinGasser/scotty/store"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	promptHeight = 3
	promptWidth  = 24
)

var (
	keyInitTyping = key.NewBinding(
		key.WithKeys(":"),
	)

	keyEnterTyping = key.NewBinding(
		key.WithKeys("enter"),
	)

	keyExitTyping = key.NewBinding(
		key.WithKeys("esc"),
	)

	keyUp = key.NewBinding(
		key.WithKeys("k"),
	)

	keyDown = key.NewBinding(
		key.WithKeys("j"),
	)

	// used to disable these keys
	// while typing in indices in the prompt
	keysTabs = []key.Binding{
		key.NewBinding(key.WithKeys("1")),
		key.NewBinding(key.WithKeys("2")),
		key.NewBinding(key.WithKeys("3")),
		key.NewBinding(key.WithKeys("4")),
	}
)

var (
	notImplemeted = lipgloss.NewStyle().
			Bold(true).
			AlignVertical(lipgloss.Center).
			AlignHorizontal(lipgloss.Center).
			Render("working on it!\n\nBrowsing logs is not yet implemented")

	defaultPromptTxt   = "type an index of a log to start browsing the logs. Use j/k to navigate up and down"
	defaultPromptChar  = "> "
	focusedPromptChar  = "> index: "
	defaultPromptStyle = lipgloss.NewStyle().Border(lipgloss.NormalBorder())
)

type Model struct {
	ready         bool
	width, height int
	bindings      bindings.Mapper
	prompt        textinput.Model
	formatter     store.Formatter
}

func New(binds bindings.Mapper, formatter store.Formatter) *Model {

	prompt := textinput.New()
	prompt.Placeholder = defaultPromptTxt
	prompt.Prompt = defaultPromptChar
	prompt.Validate = func(s string) error {
		_, err := strconv.ParseInt(s, 10, 64)
		return err
	}

	model := &Model{
		ready:     false,
		width:     0,
		height:    0,
		bindings:  binds,
		prompt:    prompt,
		formatter: formatter,
	}

	model.bindings.Map(keyInitTyping, func(msg tea.KeyMsg) tea.Cmd {
		if model.prompt.Focused() {
			return nil
		}

		model.bindings.Disable(keysTabs...)
		model.prompt.Prompt = focusedPromptChar
		return model.prompt.Focus()
	})

	model.bindings.Map(keyEnterTyping, func(msg tea.KeyMsg) tea.Cmd {
		if !model.prompt.Focused() {
			return nil
		}
		index, _ := strconv.Atoi(model.prompt.Value())
		model.formatter.Load(index)
		return nil
	})

	model.bindings.Map(keyExitTyping, func(msg tea.KeyMsg) tea.Cmd {
		model.prompt.Blur()
		model.prompt.Reset()
		model.bindings.Enable(keysTabs...)
		return nil
	})

	model.bindings.Map(keyUp, func(msg tea.KeyMsg) tea.Cmd {
		model.formatter.Privous()
		model.prompt.SetValue(strconv.Itoa(int(model.formatter.CurrentIndex())))
		return nil
	})

	model.bindings.Map(keyDown, func(msg tea.KeyMsg) tea.Cmd {
		model.formatter.Next()
		model.prompt.SetValue(strconv.Itoa(int(model.formatter.CurrentIndex())))
		return nil
	})

	return model
}

func (model Model) Init() tea.Cmd {
	return nil
}

func (model *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	var (
		cmds []tea.Cmd
		cmd  tea.Cmd
	)

	switch msg := msg.(type) {
	// case tea.KeyMsg:
	// 	switch {
	// 	// this is the initiator to start the prompt but we don't want the ":"
	// 	// to be set as a value for the prompt
	// 	case model.prompt.Focused() && key.Matches(msg, key.NewBinding(key.WithKeys(":"))):
	// 		return model, tea.Batch(cmds...)
	// 	case key.Matches(msg, model.bindings.Focus):
	// 		cmds = append(cmds,
	// 			model.prompt.Focus(),
	// 			// tell app to ignore these keys
	// 			event.BlockKeysRequest("1", "2", "3", "4"),
	// 		)
	// 		model.prompt.Prompt = focusedPromptChar
	// 	case key.Matches(msg, model.bindings.Enter):
	// 		if !model.prompt.Focused() {
	// 			break
	// 		}
	// 		// error canm be ignored as we have validation on the input prompt
	// 		index, _ := strconv.ParseInt(model.prompt.Value(), 10, 64)
	// 		model.formatter.Load(int(index))
	//
	// 	case key.Matches(msg, model.bindings.Down):
	// 		model.formatter.Next()
	// 		model.prompt.SetValue(strconv.Itoa(int(model.formatter.CurrentIndex())))
	// 	case key.Matches(msg, model.bindings.Up):
	// 		model.formatter.Privous()
	// 		model.prompt.SetValue(strconv.Itoa(int(model.formatter.CurrentIndex())))
	// 	case key.Matches(msg, model.bindings.Exit):
	// 		model.prompt.Blur()
	// 		model.prompt.Reset()
	// 		cmds = append(cmds, event.ReleaseKeysRequest())
	// 	}

	case styles.Dimensions:
		model.width = msg.Width()
		model.height = msg.Height() - promptHeight
		model.prompt.Width = promptWidth

		if !model.ready {
			model.formatter.Reset(model.width, uint8(model.height))
			model.ready = true
		}
	}

	if model.ready {
		model.prompt, cmd = model.prompt.Update(msg)
		cmds = append(cmds, cmd)
	}

	return model, tea.Batch(cmds...)
}

func (model Model) View() string {

	return lipgloss.JoinVertical(lipgloss.Left,
		defaultPromptStyle.Render(
			model.prompt.View(),
		),
		lipgloss.NewStyle().
			Height(model.height).
			Render(
				model.formatter.String(),
			),
	)
}
