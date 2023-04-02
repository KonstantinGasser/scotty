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
	// size refers to the page-size. The pager will hold
	// at-most N where is equal to `size` ring.Item(s)
	// in its buffer and serialized as raw format
	size uint8
	// width of the current tty window.
	// Mainly used to determin string break-points
	ttyWidth int
	// mode can be either "tailing" or "formatting".
	// Tailing refers to tailing the logs on MoveDown
	// while fomratting formated the a given position
	mode int
	// reader includes all required APIs
	// to perform read operations on the ringbuffer
	reader ring.Reader

	// props: mode=tailing
	//
	// position is it pagers pointer to an index in the
	// ring.Buffer. It is used to individually keep
	// track of a pagers state of tailing and allows
	// to freeze time and/or have multiple pagers with
	// different positions
	position uint32
	// buffer holds those items which are currently
	// visisble within the page - and is tight to the
	// provided size
	// buffer []ring.Item
	buffer []string
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

	// props: mode=formatting
	//
	// formatOffset if not negative refers to the an explicit
	// index in the page buffer which should be formatted
	formatOffset uint8
	// buffer filled once formatting mode is requested.
	// This buffer might not be holind the same data as
	// the tailing buffer but is filled based on the
	// requested start-index and the current page-size
	formatBuffer []ring.Item
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
	var depth int = 1
	line := fmt.Sprintf("[%d] ", next.Index()) + next.Raw[:next.DataPointer] + linewrap(
		&depth, next.Raw[next.DataPointer:],
		pager.ttyWidth-len(next.Label), len(next.Label)+3,
	)

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
		var out, height = "", 0

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
			out += item.Raw + "\n"
		}

		return out
	}

	return pager.raw
}

func (pager *Pager) Dimensions(width int, height int) {
	pager.ttyWidth = width
	pager.size = uint8(height)
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
				string(item.Label),
				string(wrap.Bytes(pretty, width-1)),
			),
		)
	return strings.Count(out, "\n"), out
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
