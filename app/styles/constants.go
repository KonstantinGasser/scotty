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
			dimensions: dimensions{
				width:  width,
				height: tabLineDefaultHeight,
			},
			style: lipgloss.NewStyle(),
		},
		Content: Content{
			dimensions: dimensions{
				width:  width,
				height: contentHeight,
			},
			style: lipgloss.NewStyle(),
		},
		FooterLine: FooterLine{
			dimensions: dimensions{
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

type dimensions struct {
	width  int
	height int
}

func (dim dimensions) Width() int {
	return dim.width
}

func (dim dimensions) Height() int {
	return dim.height
}

const (
	tabLineDefaultHeight    = 2
	footerLineDefaultHeight = 2
)

type TabLine struct {
	dimensions
	style lipgloss.Style
}

func (tabs *TabLine) Adjust(dim dimensions) {
	tabs.width = dim.width
	tabs.height = dim.height
}

func (tabs *TabLine) Render(content string) string {
	return tabs.style.Render(content)
}

type Content struct {
	dimensions
	style lipgloss.Style
}

func (content *Content) Adjust(dim dimensions) {
	content.width = dim.width
	content.height = dim.height
}

type FooterLine struct {
	dimensions
	style lipgloss.Style
}

func (footer *FooterLine) Adjust(dim dimensions) {
	footer.width = dim.width
	footer.height = dim.height
	footer.style = footer.style.Width(dim.width)
}

func (footer *FooterLine) Render(content string) string {
	return footer.style.Render(content)
}
