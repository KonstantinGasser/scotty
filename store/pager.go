package store

import (
	"fmt"
	"strings"

	"github.com/KonstantinGasser/scotty/app/styles"
	"github.com/KonstantinGasser/scotty/debug"
	"github.com/KonstantinGasser/scotty/store/ring"
	"github.com/charmbracelet/lipgloss"
)

type Pager struct {
	// reader includes all required APIs
	// to perform read operations on the ringbuffer
	reader ring.Reader
	// bufferView is the build string to display.
	// Its a representation of the buffered items
	// concatinated by a newline.
	// Note:
	// if an Item.Raw is > then the current tty width
	// the string is broken into multiple lines leading
	// to possible less items represented by the `bufferView`
	// string then present in the buffer
	bufferView string
	// buffer holds those items which are currently
	// visisble within the page - and is tight to the
	// provided size
	buffer []string
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
	// in its buffer and serialized as bufferView format
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
// The beginning of the pager's content is spliced at least
// by one (if the page is filled; if not lines are append until
// page is full). However, if the ring.Item.Raw is to wide for
// the current tty width the line is broken down in as many
// lines required. Subsequentially N lines are removed from the
// beginning of the page where N = number of broken-lines.
// The pager.bufferView is updated with the shifted content.
// MoveDown will ensure that the number of \n (lines) in the
// pager.bufferView is not exceeding the current pager.size.
func (pager *Pager) MoveDown() {

	next := pager.reader.At(pager.position)
	pager.position++

	count, lines := buildLines(next, pager.ttyWidth)

	// no issue of overflowing by adding the new lines to buffer
	if int(pager.written)+len(lines) <= int(pager.size) {
		for _, line := range lines {
			pager.buffer[pager.written] = line
			pager.written += 1
		}

		pager.bufferView = strings.Join(pager.buffer, "\n")
		return
	}

	// newly created lines will exceed the current page
	// size and we need to cut of the beginning of buffer
	pager.buffer = append(pager.buffer[count:], lines...)
	pager.bufferView = strings.Join(pager.buffer, "\n")

}

// String returns a finshed formatted string representing
// the current state of the pager.
func (pager *Pager) String() string {
	return pager.bufferView
}

// Rerender updates the pagers internal view which depends on
// the current tty width and height.
//
// Rerender flushes the current buffer and recomputes the lines
// visible within the new dimensions (width, height).
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
	pager.bufferView = strings.Join(pager.buffer, "\n")

	debug.Print("Buffer: %d Raw: %d\n", len(pager.buffer), strings.Count(pager.bufferView, "\n"))

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

// buildLines takes an ring.Item as an input and based on the
// ring.Item.Raw value and the current width of the tty returns
// a slice of string containing the broken down ring.Item.Raw line
// along with the line count.
// Example:
// 		in : label | {data: value, some: value}
// 		out: 2, ["label | {data: value,", " some: value"] for width = 21
func buildLines(item ring.Item, width int) (int, []string) {
	prefix := fmt.Sprintf("[%d] ", item.Index())
	return breaklines(
		prefix+item.Raw[:item.DataPointer],
		item.Raw[item.DataPointer:],
		width-(len(item.Label)+len(prefix)), 0,
	)
}
