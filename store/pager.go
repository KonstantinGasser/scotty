package store

import (
	"fmt"
	"strings"
	"time"

	"github.com/KonstantinGasser/scotty/store/ring"
)

type Pager struct {
	// if enabled the bufferView is
	// not updated for each received message
	paused bool
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

	writeHead int
	// ticker is used to only build the buffered view if the
	// configured refresh time has been reached in order
	// to allow to minimize the cost of re-building
	ticker *time.Ticker
}

// MovePosition moves the buffers viewing position
// by one.
// This means that any new value in the backing ring.Buffer
// is cached in the pager's buffer after being processed (broken
// into multiple lines if nessecarry).
// The buffer's view is not changed and must be called indendenly
func (pager *Pager) MovePosition() {

	next := pager.reader.At(pager.position)
	pager.position += 1

	lines := lineWrap(next, pager.ttyWidth)

	pager.shiftAppend(lines)
}

// shiftAppend takes the given lines and updates the pager's
// buffer such that the lines are append to the buffer and if
// nessecarry truncates the buffer.
func (pager *Pager) shiftAppend(lines []string) {

	var lineIndex int
	if pager.writeHead < cap(pager.buffer) {
		for pager.writeHead < cap(pager.buffer) && lineIndex < len(lines) {
			pager.buffer[pager.writeHead] = lines[lineIndex]

			pager.writeHead += 1
			lineIndex += 1
		}
	}

	var overflow = len(lines[lineIndex:])

	// no more lines to write
	if overflow <= 0 {
		return
	}

	// truncate and append buffer by N = len(lines[i:])
	// even better than truncate is to call this step:
	// shifting the buffer to the left by N.
	for i := overflow; i < cap(pager.buffer); i++ {
		pager.buffer[i-overflow] = pager.buffer[i]
	}

	shiftOffset := (cap(pager.buffer)) - overflow

	for i, j := shiftOffset, lineIndex; i < cap(pager.buffer); i, j = i+1, j+1 {
		pager.buffer[i] = lines[j]
	}
}

func (pager *Pager) PauseRender()  { pager.paused = true }
func (pager *Pager) ResumeRender() { pager.paused = false }

// String returns a finshed formatted string representing
// the current state of the pager.
func (pager *Pager) String() string {
	if pager.paused {
		return pager.bufferView
	}

	if pager.ticker == nil {
		pager.bufferView = strings.Join(pager.buffer, "\n")
		return pager.bufferView
	}

	select {
	case <-pager.ticker.C:
		pager.bufferView = strings.Join(pager.buffer, "\n")
		return pager.bufferView
	default:
		return pager.bufferView
	}
}

// Rerender updates the pagers internal view which depends on
// the current tty width and height.
//
// Rerender flushes the current buffer and recomputes the lines
// visible within the new dimensions (width, height).
func (pager *Pager) Rerender(width int, height int) {

	pager.ttyWidth = width
	pager.size = uint8(height)

	start := clamp(int(pager.position) - len(pager.buffer))
	items := make([]ring.Item, len(pager.buffer))
	pager.reader.OffsetRead(start, items)

	pager.reload(items)
	pager.bufferView = strings.Join(pager.buffer, "\n")
}

func (pager *Pager) reload(items ring.Slice) {

	pager.buffer = make([]string, pager.size)
	pager.bufferView = "Rebuilding view..."

	var written uint8
	for _, item := range items {
		if len(item.Raw) <= 0 {
			continue
		}
		lines := lineWrap(item, pager.ttyWidth)

		if int(written)+len(lines) <= int(pager.size) {
			for _, line := range lines {
				pager.buffer[written] = line
				written += 1
			}
			continue
		}

		pager.buffer = append(pager.buffer[len(lines):], lines...)
	}
}

func (pager *Pager) Reset(width int, height uint8) {
	pager.ttyWidth = width
	pager.size = height

	pager.buffer = make([]string, pager.size)
	for i := range pager.buffer {
		pager.buffer[i] = "\000"
	}
	pager.bufferView = strings.Join(pager.buffer, "\n")
}

// Refresh disregards the time.Ticker and updates
// the pager's view immediately
func (pager *Pager) Refresh() {
	pager.bufferView = strings.Join(pager.buffer, "\n")
}

func (pager *Pager) debug() string {
	return fmt.Sprintf("Height: %d\nWidth: %d\nPosition: %d, Len(buffer): %d\n", pager.size, pager.ttyWidth, pager.position, len(pager.buffer))
}

func clamp(a int) int {
	if a < 0 {
		return 0
	}
	return a
}
