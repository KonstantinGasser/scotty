package store

import (
	"fmt"
	"strings"

	"github.com/KonstantinGasser/scotty/app/styles"
	"github.com/KonstantinGasser/scotty/debug"
	"github.com/KonstantinGasser/scotty/store/ring"
	"github.com/charmbracelet/lipgloss"
	"github.com/hokaccha/go-prettyjson"
	"github.com/muesli/reflow/wrap"
)

// modes
const (
	tailing = iota
	formatting
)

type Pager struct {
	// reader includes all required APIs
	// to perform read operations on the ringbuffer
	reader ring.Reader
	// raw is the build string to display.
	// Its a representation of the buffered items
	// concatinated by a newline.
	// Note:
	// if an Item.Raw is > then the current tty width
	// the string is broken into multiple lines leading
	// to possible less items represented by the `raw`
	// string then present in the buffer
	raw string
	// buffer holds those items which are currently
	// visisble within the page - and is tight to the
	// provided size
	buffer []string
	// buffer filled once formatting mode is requested.
	// This buffer might not be holind the same data as
	// the tailing buffer but is filled based on the
	// requested start-index and the current page-size
	formatBuffer []ring.Item
	// mode can be either "tailing" or "formatting".
	// Tailing refers to tailing the logs on MoveDown
	// while fomratting formated the a given position
	// width of the current tty window.
	mode int
	// Mainly used to determin string break-points
	ttyWidth int
	// position is it pagers pointer to an index in the
	// ring.Buffer. It is used to individually keep
	// track of a pagers state of tailing and allows
	// to freeze time and/or have multiple pagers with
	// different positions
	position uint32
	// size refers to the page-size. The pager will hold
	// at-most N where is equal to `size` ring.Item(s)
	// in its buffer and serialized as raw format
	size uint8
	// written is used to intially see once the [size]buffer
	// is full since until then we only need to append data not
	// trim at the beginning
	written uint8
	// formatPosition refers to the absult (to the ring buffer) index of the item
	// which is currently formatted
	formatPosition int32
	// pageOffset refers to the index with in the formatBuffer
	// which is currently formatted
	pageOffset int8
}

// MoveDown shifts the pagers content down by one item
//
// Since a pager keeps its own offset in the underlying
// ring buffer MoveDown can be called not in sequence with
// a write to the buffer. However, one must be aware that at
// a certain point the pager's offset gets invalid once the buffer
// overflows and overflows the offset pointer of the the pager.
func (pager *Pager) MoveDown() {

	next := pager.reader.At(pager.position)
	pager.position++

	// _, line := buildLine(next, pager.ttyWidth)

	count, lines := buildLines(next, pager.ttyWidth)

	// no issue of overflowing by adding the new lines to buffer
	if int(pager.written)+len(lines) < int(pager.size) {
		for _, line := range lines {
			pager.buffer[pager.written] = line
			pager.written += 1
		}

		pager.raw = strings.Join(pager.buffer, "\n")
		return
	}

	// newly created lines will exceed the current page
	// size and we need to cut of the beginning of buffer

	pager.buffer = append(pager.buffer[count:], lines...)
	pager.raw = strings.Join(pager.buffer, "\n")

}

// EnableFormatting sets the pager in formatting mode
//
// Based on the requested start index up to the pager size
// a buffer is allocated which is independent from the
// tailing buffer.
func (pager *Pager) EnableFormatting(start uint32) {
	pager.mode = formatting

	pager.formatBuffer = pager.reader.Range(int(start-1), int(pager.size)) // negativ one to counter balance the +1 on insert
	pager.formatPosition = int32(start)
	// formatting always formates the zero (first) item
	// within the formatBuffer
	pager.pageOffset = 0
}

func (pager *Pager) FormatNext() {

	pager.pageOffset++

	if pager.pageOffset > int8(pager.size) {
		// page has turned to the next page
		debug.Print("Page turned forward I guess\n")
	}

	debug.Print("[++] Page-Size: %d Len(buff): %d - Page-Offset: %d - Format-Position: %d\n", pager.size, len(pager.formatBuffer), pager.pageOffset, pager.formatPosition)
}

func (pager *Pager) FormatPrevious() {

	pager.pageOffset--

	if pager.pageOffset <= 0 {
		// page has turned back one page
		debug.Print("Page turned back I guess\n")
		// pager.formatBuffer = pager.reader.Range(int(pager.))
	}

	debug.Print("[--] Page-Size: %d Len(buff): %d - Page-Offset: %d - Format-Position: %d\n", pager.size, len(pager.formatBuffer), pager.pageOffset, pager.formatPosition)
}

// String returns a finshed formatted string representing
// the current state of the pager.
func (pager *Pager) String() string {
	if pager.mode == formatting {

		var tmpHeight, height = 1, 1
		var tmpView, view = "", ""

		for _, item := range pager.formatBuffer {

			// normal log line which can be span multiple lines.
			// depth tells how many lines tmpView has
			tmpHeight, tmpView = buildLine(item, pager.ttyWidth)

			// adding the entire tmpView to view would overflow the
			// available space - only take as much as possible and return
			if height+tmpHeight > int(pager.size) {
				max := (height + tmpHeight) - height
				cut := strings.Split(tmpView, "\n")[:max]

				return view + strings.Join(cut, "\n")
			}

			// else we can add the entire tmpView to view
			view += tmpView + "\n"
		}

		return view
	}

	return pager.raw
}

// Rerender updates the pagers internal view which depends on
// the current tty width and height.
//
// Rerender flushes the current buffer and if filled the
// formatted buffer to recompute the displayed log lines
// based on the new provided available width and height.
func (pager *Pager) Rerender(width int, height int) {

	pager.ttyWidth = width
	pager.size = uint8(height)

	pager.buffer = pager.reader.Range(int(pager.position), int(pager.ttyWidth)).Strings(func(item ring.Item) string {
		_, line := buildLine(item, pager.ttyWidth)
		return line
	})

	if len(pager.buffer) < height {

		for i := len(pager.buffer); i < height; i++ {
			pager.buffer = append(pager.buffer, "")
		}
	}
	pager.raw = strings.Join(pager.buffer, "\n")

	debug.Print("Buffer: %d Raw: %d\n", len(pager.buffer), strings.Count(pager.raw, "\n"))

}

// overflowBy retuns the number of lines by which the
// slice of lines exceed the max height. Really just
// a function for readability...
// Is the returned number postitive, +N numbers are left
// for assignment.
// Is the returned number negative -N numbers would be to much
func overflowsBy(max uint8, lines int) int {
	return lines - int(max)
}

func breaklines(prefix string, line string, width int, padding int) (int, []string) {

	if len(line) <= width {
		return 1, []string{prefix + line}
	}

	var out []string = []string{prefix + line[:width]}
	line = line[width:]

	if len(line) <= width {
		return len(out) + 1, append(out, line)
	}

	for len(line) >= width {
		out = append(out, line[:width])

		line = line[width:]

		if len(line) <= width {
			return len(out) + 1, append(out, line)
		}
	}

	return len(out), out
}

// in : label | {data: value, some: value}
// out: 2, ["label | {data: value,", " some: value"] for width = 21
func buildLines(item ring.Item, width int) (int, []string) {
	prefix := fmt.Sprintf("[%d] ", item.Index())
	return breaklines(
		prefix+item.Raw[:item.DataPointer],
		item.Raw[item.DataPointer:],
		width-(len(item.Label)+len(prefix)), 0,
	)
}

var (
	formattedItem = lipgloss.NewStyle().
		Bold(true).
		Border(lipgloss.DoubleBorder(), false, false, false, true).
		BorderForeground(styles.DefaultColor.Border)
)

// formates the given item's raw string (only the JSON part).
// if it fails (not JSON) the raw is returns as is but only
// its data part. Alon the (formatted)string format retuns
// th approximated height of the string. Note this number is
// never less then the correct number but might be higher
// if there are escaped new line chars
func format(item ring.Item, width int) (int, string) {

	pretty, err := prettyjson.Format([]byte(item.Raw[item.DataPointer:]))
	if err != nil {
		out := formattedItem.
			Render(
				lipgloss.JoinVertical(lipgloss.Left,
					fmt.Sprintf("[%d] %s", item.Index(), item.Label),
					string(wrap.String(item.Raw[item.DataPointer:], width-1)),
				),
			)
		return strings.Count(out, "\n"), out
	}

	out := formattedItem.
		Render(
			lipgloss.JoinVertical(lipgloss.Left,
				fmt.Sprintf("[%d] %s", item.Index(), item.Label),
				string(wrap.Bytes(pretty, width-1)),
			),
		)
	return strings.Count(out, "\n"), out
}

// buildLine constructs a single line with line-breaks prefix and prefix index
//
// Example:
// [index] prefix | {data}
// [index] prefix | { data-1
// 					data-2 }
func buildLine(item ring.Item, width int) (int, string) {
	prefix := fmt.Sprintf("[%d] ", item.Index())

	height, line := linewrap(
		item.Raw[item.DataPointer:],
		width-len(item.Label)-len(prefix), 0,
	)
	return height, prefix + item.Raw[:item.DataPointer] + line
}

// linewrap breaks a line based on the given width.
// The function is not perfrect and not standard when it
// comes to line breaking however for now it serves well
// enough but is a canidate for replacement.
// Improvment could be to check if the last char is a whitespace
// and if so to remove it before adding the new line.
// Also escape seqences or ansi colors are counted as
// char which they shouldn't thou.
func linewrap(line string, width int, padding int) (int, string) {

	var height = 1
	if len(line) <= width {
		return height, line
	}

	var out = ""
	for len(line) >= width {

		out += line[:width] + "\n"
		line = line[width:]

		height += 1
		if len(line) <= width {
			return height, out + line
		}
	}

	return height, out
}
