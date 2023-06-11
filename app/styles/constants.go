package styles

import "github.com/charmbracelet/lipgloss"

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
	tabLineDefaultHeight    = 2
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
