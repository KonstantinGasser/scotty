package styles

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/ansi"
)

var (
	BgFooter = lipgloss.AdaptiveColor{
		Light: "#2c323d",
		Dark:  "#2c323d",
	}
)

type Grid struct {
	FullWidth  int
	FullHeight int

	// grid:
	// [tab] [tab] [tab]
	// .................
	// .....Content.....
	// .................
	// [footer] oX oX oX
	TabLine    TabLine
	Content    Content
	FooterLine FooterLine
}

func NewGrid(width int, height int) Grid {

	contentHeight := height - (tabLineDefaultHeight + footerLineDefaultHeight)
	return Grid{
		FullWidth:  width,
		FullHeight: height,
		TabLine: TabLine{
			Dimensions: Dimensions{
				width:  width,
				height: tabLineDefaultHeight,
			},
			style: lipgloss.NewStyle(),
		},
		Content: Content{
			Dimensions: Dimensions{
				width:  width,
				height: contentHeight,
			},
			style: lipgloss.NewStyle(),
		},
		FooterLine: FooterLine{
			Dimensions: Dimensions{
				width:  width,
				height: footerLineDefaultHeight,
			},
			style: lipgloss.NewStyle().
				MarginTop(1).
				Background(BgFooter).
				Width(width),
		},
	}
}

func (grid *Grid) Adjust(width int, height int) {
	// for tabLine and footerLine
	// the height stays static
	grid.FullWidth = width
	grid.FullHeight = height
	grid.TabLine.width = width
	grid.FooterLine.width = width
	grid.Content.width = width

	grid.Content.height = height - (grid.TabLine.height + grid.FooterLine.height)
}

type Dimensions struct {
	width  int
	height int
}

func (dim Dimensions) Width() int {
	return dim.width
}

func (dim Dimensions) Height() int {
	return dim.height
}

func (dim Dimensions) Dims() Dimensions { return dim }

const (
	tabLineDefaultHeight    = 0
	footerLineDefaultHeight = 2
)

type TabLine struct {
	Dimensions
	style lipgloss.Style
}

func (tabs *TabLine) Render(content string) string {
	return tabs.style.Render(content)
}

type Content struct {
	Dimensions
	style lipgloss.Style
}

type FooterLine struct {
	Dimensions
	style lipgloss.Style
}

func (footer *FooterLine) Render(content string) string {
	return footer.style.Render(content)
}

// SpaceBetween performs a row alignment on the left and
// right string slices where the content is pushed to the left
// and right respecifly as much as possible.
// Like in CSS display:flex; justify-content:space-between;.
// The slices should be of same length (each left[i] entry is
// matched with the right[i] entry). If the length differ, the
// short one is used as an upper limit.
func SpaceBetween(width int, left []string, right []string, betweenChar string) string {

	if width <= 0 {
		return ""
	}
	var upper = len(left)
	if len(right) < upper {
		upper = len(right)
	}

	// ignores use case where len(a) + len(b) > width
	var space = func(a string, b string) string {
		widthA := lipgloss.Width(a)
		widthB := lipgloss.Width(b)

		between := width - (widthA + widthB)
		if between < 0 {
			return ""
		}

		return strings.Join([]string{
			a,
			strings.Repeat(betweenChar, between),
			b,
		}, "")

	}

	var rows = make([]string, upper)
	for i := 0; i < upper; i++ {
		rows[i] = space(left[i], right[i])
	}

	return strings.Join(rows, "\n")
}

func FloatRight(width int, str string) string {
	if width-ansi.PrintableRuneWidth(str)-1 <= 0 {
		return str
	}

	return fmt.Sprintf(":%s%s", strings.Repeat(".", width-ansi.PrintableRuneWidth(str)-1), str)
}
