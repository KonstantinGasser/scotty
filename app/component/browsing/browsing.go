package browsing

import (
	"strconv"

	"github.com/KonstantinGasser/scotty/app/bindings"
	"github.com/KonstantinGasser/scotty/app/component/info"
	"github.com/KonstantinGasser/scotty/app/styles"
	"github.com/KonstantinGasser/scotty/store"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// we now have this in two places (app.go and browsing.go)
// maybe this should come from the info package??
type mode struct {
	label string
	bg    lipgloss.Color
}

const (
	promptHeight = 3
	promptWidth  = 24
)

/*

What if each component can create their one scope which is a bindings.Map on this scope
each component defines their scoped bindings.
The app component then could have a global scope which would be triggered globally
but then would base the KeyMsg to the currenlty active Tab Component which in tern
looks up its scoped bindings

*/

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
	bindings      *bindings.Map
	prompt        textinput.Model
	formatter     store.Formatter
}

func New(formatter store.Formatter) *Model {

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
		bindings:  bindings.NewMap(),
		prompt:    prompt,
		formatter: formatter,
	}

	model.bindings.Bind(":").
		OnESC(
			func(msg tea.KeyMsg) tea.Cmd {
				model.prompt.Blur()
				model.prompt.Reset()
				return info.RequestMode(info.ModeBrowsing)
			},
		).
		Action(
			func(msg tea.KeyMsg) tea.Cmd {
				if model.prompt.Focused() {
					return nil
				}
				model.prompt.Reset()
				model.prompt.Prompt = focusedPromptChar
				return tea.Batch(model.prompt.Focus(), info.RequestMode(info.ModePromptActive))
			},
		).
		Option("enter").Action(
		func(msg tea.KeyMsg) tea.Cmd {
			if !model.prompt.Focused() {
				return nil
			}

			index, _ := strconv.Atoi(model.prompt.Value())
			model.formatter.Load(index)
			model.prompt.Blur()

			return info.RequestMode(info.ModeBrowsing)
		})

	model.bindings.Bind("k").Action(func(msg tea.KeyMsg) tea.Cmd {
		model.formatter.Privous()
		model.prompt.SetValue(strconv.Itoa(int(model.formatter.CurrentIndex())))
		return nil
	})

	model.bindings.Bind("j").Action(func(msg tea.KeyMsg) tea.Cmd {
		model.formatter.Next()
		model.prompt.SetValue(strconv.Itoa(int(model.formatter.CurrentIndex())))
		return nil
	})

	model.bindings.Bind("r").Action(func(msg tea.KeyMsg) tea.Cmd {
		model.formatter.Load(int(model.formatter.CurrentIndex()))
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
	case styles.Dimensions:
		model.width = msg.Width()
		model.height = msg.Height() - promptHeight
		model.prompt.Width = promptWidth

		if !model.ready {
			model.ready = true
		}
		model.formatter.Resize(model.width, uint8(model.height))

	case tea.KeyMsg:
		if model.bindings.Matches(msg) {
			cmds = append(cmds, model.bindings.Exec(msg).Call(msg))
		}

	case initView:
		if !model.ready {
			break
		}
		model.formatter.Load(0)
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
