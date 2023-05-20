package styles

import (
	"bytes"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
	"github.com/muesli/ansi"
	"github.com/muesli/reflow/truncate"
	"github.com/muesli/termenv"
)

/* SCOTTY GRID LAYOUT AND STYLE */
const (
	columnOneRatio         = 0.333
	maxColumnOnWidth       = 35
	maxInfoHeight          = 2
	tabLineHeight          = 5
	ContentPaddingVertical = 1
)

func InfoWidth(width int) int {
	return int(columnOneRatio * float64(width))
}

func InfoHeight(height int) int {
	return maxInfoHeight - 2 // -2 because of borders
}

func ContentWidth(width int) int {
	return width
}

func ContentHeght(height int) int {
	return height - (TabLineHeight() + (2 * ContentPaddingVertical) + InfoHeight(height))
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
		Background(lipgloss.Color("#9F2DEB")).
		Render(label)
}

/*
BEGIN OF COPYWRITE
The following code is copied from a PR created on lipgloss
to solve the issue of overlaying elements.
Please see the PR:

	https://github.com/buztard/lipgloss/tree/feature/overlay

Could not find a Copywrite statement, however credits to @buztard (https://github.com/buztard)
who implemented and created the PR. Once the PR or an alternative is released
within the context of lipgloss this code will be replaced.

Thanks @buztard, you saved my day/ui with this one :-)!
*/
func Overlay(
	x, y int,
	fg, bg string,
	shadow bool,
) string {
	fgLines, fgWidth := getLines(fg)
	bgLines, bgWidth := getLines(bg)
	bgHeight := len(bgLines)
	fgHeight := len(fgLines)

	if shadow {
		var shadowbg string = ""
		shadowchar := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#333333")).
			Render("░")
		for i := 0; i <= fgHeight; i++ {
			if i == 0 {
				shadowbg += " " + strings.Repeat(" ", fgWidth) + "\n"
			} else {
				shadowbg += " " + strings.Repeat(shadowchar, fgWidth) + "\n"
			}
		}

		fg = Overlay(0, 0, fg, shadowbg, false)
		fgLines, fgWidth = getLines(fg)
		fgHeight = len(fgLines)
	}

	if fgWidth >= bgWidth && fgHeight >= bgHeight {
		// FIXME: return fg or bg?
		return fg
	}
	// TODO: allow placement outside of the bg box?
	x = clamp2(x, 0, bgWidth-fgWidth)
	y = clamp2(y, 0, bgHeight-fgHeight)

	ws := &whitespace{}

	var b strings.Builder
	for i, bgLine := range bgLines {
		if i > 0 {
			b.WriteByte('\n')
		}
		if i < y || i >= y+fgHeight {
			b.WriteString(bgLine)
			continue
		}

		pos := 0
		if x > 0 {
			left := truncate.String(bgLine, uint(x))
			pos = ansi.PrintableRuneWidth(left)
			b.WriteString(left)
			if pos < x {
				b.WriteString(ws.render(x - pos))
				pos = x
			}
		}

		fgLine := fgLines[i-y]
		b.WriteString(fgLine)
		pos += ansi.PrintableRuneWidth(fgLine)

		right := cutLeft(bgLine, pos)
		bgWidth := ansi.PrintableRuneWidth(bgLine)
		rightWidth := ansi.PrintableRuneWidth(right)
		if rightWidth <= bgWidth-pos {
			b.WriteString(ws.render(bgWidth - rightWidth - pos))
		}

		b.WriteString(right)
	}

	return b.String()
}

type whitespace struct {
	style termenv.Style
	chars string
}

func getLines(s string) (lines []string, widest int) {
	lines = strings.Split(s, "\n")

	for _, l := range lines {
		w := ansi.PrintableRuneWidth(l)
		if widest < w {
			widest = w
		}
	}

	return lines, widest
}

// cutLeft cuts printable characters from the left.
// This function is heavily based on muesli's ansi and truncate packages.
func cutLeft(s string, cutWidth int) string {
	var (
		pos    int
		isAnsi bool
		ab     bytes.Buffer
		b      bytes.Buffer
	)
	for _, c := range s {
		var w int
		if c == ansi.Marker || isAnsi {
			isAnsi = true
			ab.WriteRune(c)
			if ansi.IsTerminator(c) {
				isAnsi = false
				if bytes.HasSuffix(ab.Bytes(), []byte("[0m")) {
					ab.Reset()
				}
			}
		} else {
			w = runewidth.RuneWidth(c)
		}

		if pos >= cutWidth {
			if b.Len() == 0 {
				if ab.Len() > 0 {
					b.Write(ab.Bytes())
				}
				if pos-cutWidth > 1 {
					b.WriteByte(' ')
					continue
				}
			}
			b.WriteRune(c)
		}
		pos += w
	}
	return b.String()
}

func clamp2(v, lower, upper int) int {
	return min(max(v, lower), upper)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (w whitespace) render(width int) string {
	if w.chars == "" {
		w.chars = " "
	}

	r := []rune(w.chars)
	j := 0
	b := strings.Builder{}

	// Cycle through runes and print them into the whitespace.
	for i := 0; i < width; {
		b.WriteRune(r[j])
		j++
		if j >= len(r) {
			j = 0
		}
		i += ansi.PrintableRuneWidth(string(r[j]))
	}

	// Fill any extra gaps white spaces. This might be necessary if any runes
	// are more than one cell wide, which could leave a one-rune gap.
	short := width - ansi.PrintableRuneWidth(b.String())
	if short > 0 {
		b.WriteString(strings.Repeat(" ", short))
	}

	return w.style.Styled(b.String())
}

/* END OF COPYWRITE */
