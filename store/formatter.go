package store

import (
	"strings"

	"github.com/KonstantinGasser/scotty/app/styles"
	"github.com/KonstantinGasser/scotty/store/ring"
	"github.com/charmbracelet/lipgloss"
	"github.com/hokaccha/go-prettyjson"
	"github.com/muesli/ansi"
	"github.com/muesli/reflow/wrap"
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

	// absolute refers to the ring.Item.index which
	// is currently absolute and should be formatted
	absolute uint32
	// relative refers to the index offeset relativ to
	// the current buffer/page and gives indications
	// about page turns (froward and backend)
	relative uint8
	// visibleItemCount is the actual number of items
	// which can currently be displayed under the current
	// tty dimensions. It is given that visibleItemCount
	// will never be > the size/len(buffer)
	visibleItemCount uint8
	// mainly used for worwrapping
	ttyWidth int
}

func (formtter Formatter) CurrentIndex() uint32 {
	return formtter.buffer[formtter.relative].Index()
}

func (formatter *Formatter) Load(start int) {

	formatter.buffer = formatter.reader.Range(start, int(formatter.size))

	formatter.relative = 0 // make the first item of the buffer be the absolute item
	formatter.absolute = uint32(start)

	formatter.buildView()
}

func (formatter *Formatter) Next() {

	formatter.absolute += 1

	// turn page forward by formatter.size
	if formatter.relative+1 > formatter.size {
		formatter.buffer = formatter.reader.Range(int(formatter.absolute), int(formatter.size))
		formatter.relative = 0

		formatter.buildView()
		return
	}

	formatter.relative += 1
	formatter.buildView()
}

func (formatter *Formatter) Privous() {

	if formatter.absolute-1 == 0 {
		return
	}

	formatter.absolute -= 1
	if formatter.relative == 0 {
		formatter.buffer = formatter.reader.Range(int(formatter.absolute), int(formatter.size))

		formatter.buildView()
		return
	}

	formatter.relative -= 1
	formatter.buildView()
}

// assumes that the buffer has the correct data
func (formatter *Formatter) buildView() {
	formatter.buildBackground()
	formatter.buildForeground()
}

var selected = ">>>"
var trimmedSuffix = "..."

func (formatter *Formatter) buildBackground() {

	var lines = make([]string, formatter.size)

	var printable, ansiEsc int
	for i, item := range formatter.buffer {

		var raw = strings.Builder{}

		if i == int(formatter.relative) {
			raw.WriteString(selected)
		}
		raw.WriteString(item.Raw)

		printable = ansi.PrintableRuneWidth(raw.String())
		ansiEsc = len(raw.String()) - printable

		if printable > formatter.ttyWidth {
			tmp := raw.String()[:formatter.ttyWidth+(ansiEsc-1)]
			lines[i] = tmp[:len(tmp)-1-len(trimmedSuffix)] + trimmedSuffix
			continue
		}

		lines[i] = raw.String()
		raw.Reset()
	}
	formatter.background = strings.Join(lines, "\n")
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

	item := formatter.reader.At(uint32(formatter.absolute))

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
	formatter.absolute = 0
	formatter.buffer = make([]ring.Item, formatter.size)
}

func (formatter Formatter) String() string {

	modalWidth, modalHeight := lipgloss.Width(formatter.foreground), lipgloss.Height(formatter.foreground)
	x0 := int(formatter.ttyWidth/2) - int(modalWidth/2)
	y0 := int(formatter.size/2) - int(modalHeight/2)

	return styles.Overlay(x0, y0, formatter.foreground, formatter.background, false)
}
