package styles

import (
	"github.com/charmbracelet/lipgloss"
)

/* SCOTTY GRID LAYOUT AND STYLE */
const (
	columnOneRatio         = 0.333
	maxColumnOnWidth       = 25
	tabLineHeight          = 3
	ContentPaddingVertical = 1
)

func InfoWidth(width int) int {
	return max(maxColumnOnWidth, int(columnOneRatio*float64(width)))
}

func ContentWidth(width int) int {
	return width - InfoWidth(width)
}

func ContentHeght(height int) int {
	return height - (TabLineHeight() + (2 * ContentPaddingVertical))
}

func TabLineHeight() int {
	return tabLineHeight
}

/* SCOTTY TAB LAYOUT AND STYLE  */
var (
	tab = lipgloss.NewStyle().
		Padding(0, 1).
		MarginRight(1).
		Background(lipgloss.Color("#c4c4c4")).
		Foreground(lipgloss.Color("#ffffff"))
)

func Tab(label string) string {
	return tab.Render(label)
}

func ActiveTab(label string) string {
	return tab.Copy().
		Bold(true).
		Background(lipgloss.Color("#FF4C94")).
		Render(label)
}

func max(upper int, compare int) int {
	if compare > upper {
		return upper
	}
	return compare
}
