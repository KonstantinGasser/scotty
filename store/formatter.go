package store

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/KonstantinGasser/scotty/store/ring"
	"github.com/charmbracelet/lipgloss"
	"github.com/hokaccha/go-prettyjson"
	"github.com/mattn/go-runewidth"
	"github.com/muesli/ansi"
	"github.com/muesli/reflow/truncate"
	"github.com/muesli/reflow/wrap"
	"github.com/muesli/termenv"
)

var (
	jsonF = prettyjson.NewFormatter()
)

func init() {
	jsonF.StringColor = nil
}

type Formatter struct {
	reader ring.Reader
	// page size - max number of items
	// which can be placed on the page
	// without any of them being formatted.
	size uint8
	// buffer of the items currently visiable
	// in the view.
	buffer []ring.Item
	view   string

	background string
	foreground string

	// selected refers to the ring.Item.index which
	// is currently selected and should be formatted
	selected uint32
	// offset refers to the index offeset relativ to
	// the current buffer/page and gives indications
	// about page turns (froward and backend)
	offset uint8
	// visibleItemCount is the actual number of items
	// which can currently be displayed under the current
	// tty dimensions. It is given that visibleItemCount
	// will never be > the size/len(buffer)
	visibleItemCount uint8
	// mainly used for worwrapping
	ttyWidth int
}

func (formatter *Formatter) Load(start int) {

	formatter.buffer = make([]ring.Item, formatter.size)
	formatter.reader.OffsetWrite(start, formatter.buffer)

	formatter.offset = 0 // make the first item of the buffer be the selected item
	formatter.selected = uint32(start - 1)

	formatter.buildView()
}

func (formatter *Formatter) Next() {

	formatter.selected += 1

	// turn page forward by formatter.size
	if formatter.offset+1 > formatter.visibleItemCount-1 {
		formatter.buffer = formatter.reader.Range(int(formatter.selected), int(formatter.size))
		formatter.offset = 0

		formatter.buildView()
		return
	}

	formatter.offset += 1
	formatter.buildView()
}

func (formatter *Formatter) Privous() {

	if formatter.selected-1 == 0 {
		return
	}

	formatter.selected -= 1
	if formatter.offset == 0 {
		formatter.buffer = formatter.reader.Range(int(formatter.selected), int(formatter.size))

		formatter.buildView()
		return
	}

	formatter.offset -= 1
	formatter.buildView()
}

// assumes that the buffer has the correct data
func (formatter *Formatter) buildView() {
	formatter.buildBackground()
	formatter.buildForeground()
}

func (formatter *Formatter) buildBackground() {

	var height, lines, tmp = 0, []string{}, make([]string, formatter.size)
	var written uint8

	formatter.visibleItemCount = 0

	for i, item := range formatter.buffer {

		if written >= formatter.size {
			break
		}

		formatter.visibleItemCount += 1

		lines = nil
		if len(item.Raw) <= 0 {
			continue
		}

		var prefixOptions []func(string) string
		if i == int(formatter.offset) {
			prefixOptions = append(prefixOptions, func(s string) string {
				return fmt.Sprintf("%s%s", lipgloss.NewStyle().Bold(true).Render(">>"), s)
			})
		}

		height, lines = buildLines(item, formatter.ttyWidth, prefixOptions...)

		if int(written)+height <= int(formatter.size) {
			for _, line := range lines {
				tmp[written] = line
				written += 1
			}
			continue
		}

		tmp = append(tmp[height:], lines...)
	}

	formatter.background = strings.Join(tmp, "\n")
}

var (
	modalStyle = lipgloss.NewStyle().
			Padding(0, 1).
			Border(lipgloss.RoundedBorder())
	modalInfoStyle = lipgloss.NewStyle().
			Bold(true).
			Border(lipgloss.DoubleBorder(), false, false, true, false)
)

const (
	modalWidthRatio = 0.8
)

func modalWidth(full int) int {
	return int(float64(full) * (modalWidthRatio))
}

func (formatter *Formatter) buildForeground() {

	item := formatter.reader.At(uint32(formatter.selected))

	pretty, _ := jsonF.Format(
		[]byte(item.Raw[item.DataPointer:]),
	)

	broken := wrap.Bytes(pretty, modalWidth(formatter.ttyWidth))

	content := lipgloss.JoinVertical(lipgloss.Left,
		item.Raw[:item.DataPointer],
		string(broken),
	)

	formatter.foreground = modalStyle.
		Width(modalWidth(formatter.ttyWidth)).
		Render(content)
}

func (formatter *Formatter) Reset(width int, height uint8) {
	formatter.ttyWidth = width
	formatter.size = height
	formatter.selected = 0
	formatter.buffer = make([]ring.Item, formatter.size)
}

func (formatter Formatter) String() string {

	modalWidth, modalHeight := lipgloss.Width(formatter.foreground), lipgloss.Height(formatter.foreground)
	x0 := int(formatter.ttyWidth/2) - int(modalWidth/2)
	y0 := int(formatter.size/2) - int(modalHeight/2)

	return PlaceOverlay(x0, y0, formatter.foreground, formatter.background, false)
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
func PlaceOverlay(
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

		fg = PlaceOverlay(0, 0, fg, shadowbg, false)
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
