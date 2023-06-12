package store

import (
	"fmt"
	"strings"
	"time"

	"github.com/KonstantinGasser/scotty/store/ring"
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

	ticker *time.Ticker
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
func (pager *Pager) MoveDown(skipRefresh bool) {
	next := pager.reader.At(pager.position)
	pager.position++
	// lines holds a single log line wrapped
	// into multiple strings each no longer than
	// pager.ttyWidth
	_, lines := buildLines(next, pager.ttyWidth)

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
		height, lines := buildLines(item, pager.ttyWidth)

		if int(written)+height <= int(pager.size) {
			for _, line := range lines {
				pager.buffer[written] = line
				written += 1
			}
			continue
		}

		pager.buffer = append(pager.buffer[height:], lines...)
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
