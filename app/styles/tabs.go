package styles

import "github.com/charmbracelet/lipgloss"

var (
	fgDefault = lipgloss.AdaptiveColor{
		Light: "43",
		Dark:  "43",
	}

	fgActive = lipgloss.AdaptiveColor{
		Light: "0",
		Dark:  "0",
	}

	tabDefaultStyle = lipgloss.NewStyle().
			MarginRight(1).
			Foreground(fgDefault)

	tabActiveStyle = tabDefaultStyle.Copy().
			Foreground(fgActive).
			Background(lipgloss.Color("43"))
)

type Tabs struct {
	lables []string
	active int
	view   string
}

func NewTabs(active int, lables ...string) *Tabs {
	tabs := &Tabs{
		lables: lables,
		active: active,
		view:   "",
	}

	tabs.build()

	return tabs
}

func (tabs *Tabs) SetActive(index int) {
	if index > len(tabs.lables)-1 {
		return
	}

	tabs.active = index
	tabs.build()
}

func (tabs *Tabs) build() {
	var items = make([]string, len(tabs.lables))
	for i := range items {
		if i == tabs.active {
			items[i] = tabActiveStyle.Render(tabs.lables[i])
			continue
		}
		items[i] = tabDefaultStyle.Render(tabs.lables[i])
	}

	tabs.view = lipgloss.JoinHorizontal(lipgloss.Left, items...)
}

var (
	tabStyle = lipgloss.NewStyle().PaddingBottom(1)
)

var (
	tabsStyle = lipgloss.NewStyle().PaddingBottom(1)
)

func (tabs Tabs) View() string {
	return tabsStyle.
		Render(tabs.view)
}
