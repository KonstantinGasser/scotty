package store

import (
	"strings"

	"github.com/KonstantinGasser/scotty/store/ring"
)

type Pager struct {
	// size refers to the page-size. The pager will hold
	// at-most N where is equal to `size` ring.Item(s)
	// in its buffer and serialized as raw format
	size uint8
	// width of the current tty window.
	// Mainly used to determin string break-points
	ttyWidth int
	// reader includes all required APIs
	// to perform read operations on the ringbuffer
	reader ring.Reader
	// position is it pagers pointer to an index in the
	// ring.Buffer. It is used to individually keep
	// track of a pagers state of tailing and allows
	// to freeze time and/or have multiple pagers with
	// different positions
	position uint32
	// buffer holds those items which are currently
	// visisble within the page - and is tight to the
	// provided size
	buffer []ring.Item
	// written is used to intially see once the [size]buffer
	// is full since until then we only need to append data not
	// trim at the beginning
	written uint8
	// raw is the build string to display.
	// Its a representation of the buffered items
	// concatinated by a newline.
	// Note:
	// if an Item.Raw is > then the current tty width
	// the string is broken recursivly leading to possible less items
	// represented by the `raw` string then present in the buffer
	raw string
}

// MoveDown shifts the pagers content down by one item
//
// It should be called whenever a new Item is inserted
// into the store.Store.buffer.
// This function is most likely called frequentially
// as such re-computing each state again is a wast of
// CPU and memory since to move down really only means
// that the start and the end of the buffer/raw-string
// change.
// To address this MoveDonw strips away the first line
// in the raw-string and only adds the new line.
// Similar for the buffer. The zero item of the buffer
// is discarded while the new one is added at the end.
func (pager *Pager) MoveDown() {

	next := pager.reader.At(pager.position)
	pager.position++

	// actual height of the resulting string
	var depth int
	line := linewrap(&depth, next.Raw, pager.ttyWidth)

	// filling up the buffer before we can start
	// windowing
	if pager.written < pager.size {
		pager.buffer[int(pager.written)] = next
		pager.raw += line

		pager.written++
		return
	}

	// move the window one to the right
	// for both the buffer and the raw string
	pager.buffer = pager.buffer[1:] // cutof first value of the buffer
	pager.buffer[len(pager.buffer)-1] = next
}

func shiftString(base string, line string, height int) string {
	// for now we ignore that any line where the height is
	// > 1 implies that the pager's raw string is higher then
	// the pager's actual size

	cut := strings.IndexByte(base, byte('\n'))
	if cut < 0 {
		return base + line
	}
	return base[cut+1:] + line
}

// linewrap breaks a line based on the given width.
// The function is not perfrect and not standard when it
// comes to line breaking however for now it serves well
// enough but is a canidate for replacement.
// Improvment could be to check if the last char is a whitespace
// and if so to remove it before adding the new line.
func linewrap(depth *int, line string, width int) string {
	if len(line) <= width {
		return line
	}

	*depth = (*depth) + 1
	return line[:width] + "\n" + linewrap(depth, line[width:], width)
}

func (pager *Pager) String() string {
	return pager.raw
}

func (pager *Pager) Dimensions(width int, height int) {
	pager.ttyWidth = width
	pager.size = uint8(height)
}
