package store

import (
	"fmt"
	"strings"

	"github.com/KonstantinGasser/scotty/app/styles"
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
	// the string is broken recursivly leading to possible less items
	// represented by the `raw` string then present in the buffer
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
	// formatOffset if not negative refers to the an explicit
	// index in the page buffer which should be formatted
	formatOffset uint8
}

// MoveDown shifts the pagers content down by one item
//
// Since a pager keeps its own offset in the underlying
// ring buffer MoveDown can be called not in sequence with
// a write to the buffer. However, one must be aware that at
// a certain point the pager's offset gets invalid once the buffer
// overflows and overflows the offset pointer of the the pager.
//
// Dreamworld...not yet implemented:
// This cannot be detected by the MoveDown function but can be
// checked calling pager.IsBadRead() bool. If true a call to
// pager.Latest() sets the pager to the latest item in the buffer.
func (pager *Pager) MoveDown() {

	next := pager.reader.At(pager.position)
	pager.position++

	// actual height of the resulting string
	// TODO @KonstantinGasser:
	// eventually we need to cut of more lines
	// at the beginnin if the combined depth/height
	// is > pager.size...yet to be implemented
	var depth int = 1
	line := buildLine(&depth, next, pager.ttyWidth)

	// filling up the buffer before we can start
	// windowing
	if pager.written < pager.size {
		// next.Parsed = line
		pager.buffer[int(pager.written)] = line

		pager.written++
		pager.raw = strings.Join(pager.buffer[:pager.written], "\n")

		return
	}

	pager.buffer = append(pager.buffer[1:], line)
	pager.raw = strings.Join(pager.buffer, "\n")
}

// linewrap breaks a line based on the given width.
// The function is not perfrect and not standard when it
// comes to line breaking however for now it serves well
// enough but is a canidate for replacement.
// Improvment could be to check if the last char is a whitespace
// and if so to remove it before adding the new line.
// Also escape seqences or ansi colors are counted as
// char which they shouldn't thou.
func linewrap(depth *int, line string, width int, padding int) string {

	if width <= 0 || len(line) <= width {
		if *depth > 1 {
			return strings.Repeat(" ", padding) + line
		}
		return line
	}

	*depth = (*depth) + 1

	return line[:width] + "\n" + linewrap(depth, line[width:], width, padding)
}

// EnableFormatting sets the pager in formatting mode
//
// Based on the requested start index up to the pager size
// a buffer is allocated which is independent from the
// tailing buffer.
func (pager *Pager) EnableFormatting(start uint32) {
	pager.mode = formatting

	pager.formatBuffer = pager.reader.Range(int(start-1), int(pager.size)) // negativ one to counter balance the +1 on insert
	pager.formatOffset = 0
}

func (pager *Pager) FormatNext() {

	if pager.formatOffset > pager.size {
		// turn page forward
		return
	}

	pager.formatOffset++
}

func (pager *Pager) FormatPrevious() {

	if pager.formatOffset < 0 {
		// turn page back
		return
	}

	pager.formatOffset--
}

// String returns a finshed formatted string representing
// the current state of the pager.
func (pager *Pager) String() string {
	if pager.mode == formatting {
		var out, height, depth = "", 0, 0

		for i, item := range pager.formatBuffer {
			if height >= int(pager.size) {
				return out
			}

			if int(pager.formatOffset) == i {
				h, f := format(item, pager.ttyWidth)
				out += f + "\n"
				height += h
				continue
			}
			prefix := fmt.Sprintf("[%d] %s", item.Index(), item.Label)
			out += prefix + linewrap(&depth, item.Raw[item.DataPointer:], pager.ttyWidth, len(prefix)) + "\n"
		}

		return out
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
	// pager.ttyWidth = width
	// pager.size = uint8(height)

	// both modes:
	// if the height changes we need to do more

	// implies that the current buffer and if set
	// formatted buffer are invalid and either contain
	// to little or to much items - we need to build the
	// buffer starting from its position back to position-height
	if int(pager.size) != height {
		var depth int
		tmp := pager.reader.Range(int(pager.position), height).Strings(func(i ring.Item) string {
			return buildLine(&depth, i, width) // improtant to take the provided with if width != pager.ttyWidth
		})
		pager.raw = strings.Join(tmp, "\n")
	}
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
func buildLine(depth *int, item ring.Item, width int) string {
	return fmt.Sprintf("[%d] ", item.Index()) + item.Raw[:item.DataPointer] + linewrap(
		depth, item.Raw[item.DataPointer:],
		width-len(item.Label), len(item.Label)+3,
	)
}

// deprecated as of now
func shiftString(base string, line string, height int) string {
	// for now we ignore that any line where the height is
	// > 1 implies that the pager's raw string is higher then
	// the pager's actual size

	cut := strings.IndexByte(base, byte('\n'))
	if cut < 0 {
		return base + line + "\n"
	}

	return base[cut+1:] + line + "\n"
}
