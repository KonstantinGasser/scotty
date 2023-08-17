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
	// scrollDelta is used to determine the delta position
	// when starting to scroll up or down. Scrolling is only
	// possible if the tailing mode is paused (we can debate
	// if it makes sense to pause tailing the moment a ScrollMsg
	// is emitted).
	// A negative delta means based on the starting position
	// content above/previous from that position is requested.
	// A delta of 0 implies that the scrolled view matches the
	// initital view when the tailing has been paused
	scrollDelta int32
	// scrollBuffer holds the current items which are
	// determined by the scrollDelta. This buffer is
	// only filled and maintained while scrolling is
	// executed afterwards its drained.
	scrollBuffer []ring.Item
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
//
// Major issue as of now:
// The capacity of the buffer is increasing to most twice
// the starting capacity of the buffer and then slowly shrinks
// back to the starting capacity which should increase the
// memmove and growSlice runtime calls which have a certain
// performance penalty.
// This issue will not be fixed in this function but rather a
// new improved function is written.
func (pager *Pager) MoveDownDeprecated(skipRefresh bool) {
	// defer debug.Print("Len: %d - Cap: %d\n", len(pager.buffer), cap(pager.buffer))

	next := pager.reader.At(pager.position)
	pager.position++
	// lines holds a single log line wrapped
	// into multiple strings each no longer than
	// pager.ttyWidth
	lines := lineWrap(next, pager.ttyWidth)
	// debug.Print("Len: %d - Cap: %d - NewLines: %d\n", len(pager.buffer), cap(pager.buffer), len(lines))

	// insert new log line into buffer
	for _, line := range lines {
		// while buffer is not full (number of logs written
		// is less than the page size) we can write into the
		// next index
		if int(pager.written) < int(pager.size) {
			pager.buffer[pager.written] = line
			pager.written += 1
			continue
		}
		// default behaviour:
		// cut first element of buffer and append new line
		// since the cap never changes this operation should
		// be efficient enough
		pager.buffer = append(pager.buffer[1:], line)
	}

	if skipRefresh {
		return
	}

	if pager.ticker == nil {
		pager.bufferView = strings.Join(pager.buffer, "\n")
		return
	}

	// only update the buffer view if refresh rate ticks
	select {
	case <-pager.ticker.C:
		pager.bufferView = strings.Join(pager.buffer, "\n")
	default: // empty default required else waiting for refresh and causing buffering of messages

	}
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

	// at this point we have N new lines we need to append in the buffer.
	// for this to work we need to consider the fact that the buffer has
	// a fixed cap and len and is inititally filled with \000 values.
	// one requirement is to not change the the buffers cap causing
	// runtime.growslice calls leading to overhead work.
	// as such the first thing we need to do is to check how full the buffer
	// currently is. For this we can use the pager.written value which indicates
	// how many elements have been written into the buffer so far.
	// Next we need to fill the buffer until full. (Caution: there might be more
	// lines to write then we can put in the not full buffer).
	// Once the buffer is full we need to truncate N elements from the beginning
	// of the buffer (where N is len(lines)) and append all lines at the end of the
	// buffer. Here we need to be cautious about not causing a growslice call by
	// changing the buffer's capacity.

	// this means the buffer has still free (pre-initialised) slots
	// of \000 values which we need to overwrite before starting to
	// truncate the beginning of the buffer.

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

func (pager Pager) debugf(msg string, args ...interface{}) {
	defer fmt.Println("====END===")
	fmt.Println("====DEBUG====")
	fmt.Printf(msg, args...)
	for _, line := range pager.buffer {
		fmt.Println(line)
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
	pager.bufferView = "Rebulding view..."

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
		pager.written = written
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
