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
