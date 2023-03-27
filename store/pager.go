package store

import "github.com/KonstantinGasser/scotty/store/ring"

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

	// next := pager.reader.At(pager.position)

}

func linewrap(depth *int, line string, width int) string {
	if len(line) <= width {
		return line
	}

	*depth = (*depth) + 1
	return line[:width] + "\n" + linewrap(depth, line[width:], width)
}
