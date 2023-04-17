package store

import (
	"strings"

	"github.com/KonstantinGasser/scotty/debug"
	"github.com/KonstantinGasser/scotty/store/ring"
	"github.com/charmbracelet/lipgloss"
	"github.com/hokaccha/go-prettyjson"
	"github.com/muesli/reflow/wrap"
)

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
	// selected refers to the ring.Item.index which
	// is currently selected and should be formatted
	selected uint32
	// mainly used for worwrapping
	ttyWidth int
}

func (formatter *Formatter) Init(start int) {
	formatter.buffer = formatter.reader.Range(start-1, int(formatter.size))
	for _, i := range formatter.buffer {
		debug.Print("| %d | ", i.Index())
	}
	debug.Print("\n")
	formatter.selected = uint32(start)
	formatter.buildView()
}

func (formatter *Formatter) Next() {
	formatter.selected++
	formatter.buildView()
}

func (formatter *Formatter) Privous() {
	formatter.selected--
	formatter.buildView()
}

// assumes that the buffer has the correct data
func (formatter *Formatter) buildView() {

	var height, lines, tmp = 0, []string{}, make([]string, formatter.size)
	var written uint8

	for _, item := range formatter.buffer {
		if len(item.Raw) <= 0 {
			continue
		}
		debug.Print("[formatter] Index: %d Selected: %d\n", item.Index(), formatter.selected)
		if item.Index() == formatter.selected {
			pretty, _ := prettyjson.Format([]byte(item.Raw[item.DataPointer:]))
			wrapped := wrap.Bytes(pretty, formatter.ttyWidth)

			lines = append(lines, item.Raw[:item.DataPointer])
			height, lines = lipgloss.Height(string(wrapped)), append(lines, strings.Split(string(wrapped), "\n")...)

		} else {
			height, lines = buildLines(item, formatter.ttyWidth)
		}

		if int(written)+height <= int(formatter.size) {
			for _, line := range lines {
				tmp[written] = line
				written += 1
			}
			continue
		}
		tmp = append(tmp[height:], lines...)
	}

	formatter.view = strings.Join(tmp, "\n")
}

func (formatter *Formatter) Reset(width int, height uint8) {
	formatter.ttyWidth = width
	formatter.size = height
	formatter.selected = 0
	formatter.buffer = make([]ring.Item, formatter.size)
}

func (formatter Formatter) String() string {
	return formatter.view
}
