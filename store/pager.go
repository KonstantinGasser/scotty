package store

import (
	"fmt"
	"strings"
	"time"

	"github.com/KonstantinGasser/scotty/debug"
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

	debug.Print("Lines:\n%s\n", strings.Join(lines, "\n"))
	// if less than zero -> pager can write all lines in the not yet full
	// buffer.
	// Results > 0 imply that only so many lines can be written in the not full
	// buffer before the starting items must be truncated
	overflow := (int(pager.written) + len(lines)) - cap(pager.buffer) // cap: 20 written: 15 len(lines): 10 => 15+10 - 20 => 25 - 20 => 5

	if overflow <= 0 {
		for _, line := range lines {
			pager.buffer[pager.written] = line
			pager.written += 1
		}

		// we are done all possible lines where written
		// in the buffer
		return
	}

	// TODO:
	// in pager.reload we need to update the pager.written value..I think at least
	// we get a panic after buffer full -> width resize of terminal

	// ok at this point we know that either not all lines fitted
	// in the not yet full buffer and some are left-over or that the
	// buffer was full to begin with an all lines must be append and
	// the first lines in the buffer must be truncated.
	// Either way it's the same...

	// shift the buffer to the left by how many lines need to be written
	// essentially truncating N lines from the beginning of the buffer.
	// Question:
	// why not using sliceing [N:], see func MoveDownDeprecated and
	// the not regarding growSlice and memmove.
	for i := overflow; i < cap(pager.buffer); i++ {
		pager.buffer[i-overflow] = pager.buffer[i]
	}

	// append new lines without increasing the slice capacity
	for i, j := (cap(pager.buffer)-1)-(overflow-1), 0; i < cap(pager.buffer); i, j = i+1, j+1 {
		pager.buffer[i] = lines[j]
	}

	// fmt.Println("=====START=====")
	// fmt.Println(strings.Join(pager.buffer, "\n"))
	// fmt.Println("=====END=====")
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
