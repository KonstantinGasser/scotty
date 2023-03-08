package ring

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/KonstantinGasser/scotty/app/styles"
	"github.com/KonstantinGasser/scotty/ring/filter"
	"github.com/charmbracelet/lipgloss"
	"github.com/hokaccha/go-prettyjson"
	"github.com/muesli/reflow/wrap"
)

type Log struct {
	Label string
	Data  []byte
}

type Buffer struct {
	capacity uint32
	write    uint32
	data     []Log
	filter   *filter.Filter
}

// New initiates a new ring buffer with a set capacity.
// To calculate the approximated memory size
// one has to take the size of the on average expected []byte
// stored in the buffer and compute: capacity*avg(item_size)
func New(size uint32) Buffer {
	return Buffer{
		capacity: size,
		write:    0,
		data:     make([]Log, size),
		filter:   filter.New(filter.Default),
	}
}

func (buf Buffer) Cap() uint32 {
	return buf.capacity
}

func (buf *Buffer) ApplyFilter(fn filter.FilterFunc, fields ...string) {
	buf.filter = filter.New(fn, fields...)
}

func (buf *Buffer) AddFilter(fields ...string) {
	buf.filter.Append(fields...)
}

func (buf *Buffer) RemoveFilter(field string) {
	buf.filter.Remove(field)
}

func (buf *Buffer) UnsetFilter() {
	buf.filter = filter.New(filter.Default)
}

func (buff Buffer) Nil(index int) bool {
	return len(buff.data[index].Data) == 0
}

var (
	ErrReadBeforeWrite = fmt.Errorf("requested index has not been written yet - empty data")
)

func (buf Buffer) TryRead(index int) (bool, error) {
	if index > int(buf.capacity) || index < 0 {
		return false, ErrIndexOutOfBounds
	}

	if len(buf.data[index].Data) <= 0 {
		return false, ErrReadBeforeWrite
	}

	return buf.filter.Test(buf.data[index].Label, buf.data[index].Data), nil
}

func (buf *Buffer) Write(label string, data []byte) (int, error) {
	buf.data[buf.write].Label = label
	// previously the index was append on each read.
	// However, scotty does more iterations reading from the buffer
	// then writing. As such it is more efficient to append the index on write
	buf.data[buf.write].Data = append([]byte("["+strconv.Itoa(int(buf.write))+"]"), data...)

	buf.write = (buf.write + 1) % buf.capacity
	return len(label) + len(data), nil
}

func (buf *Buffer) Read(w *bytes.Buffer, rangeN int, fns ...func(int, []byte) []byte) (int, error) {

	var written int

	offset := int(buf.write) - rangeN
	cap := int(buf.capacity)

	var b []byte

	var index int
	for i := offset; i < int(buf.write); i++ {

		index = (cap - 1) - ((((-i - 1) + cap) % cap) % cap)

		if len(buf.data[index].Data) == 0 {
			continue
		}

		if ok := buf.filter.Test(buf.data[index].Label, buf.data[index].Data); !ok {
			continue
		}

		b = buf.data[index].Data
		for _, fn := range fns {
			b = fn(index, b)
		}

		if _, err := w.Write(b); err != nil {
			return written, err
		}
		written++
	}
	return written, nil
}

func (buf *Buffer) ReadOffset(w *bytes.Buffer, offset int, rangeN int, fns ...func(int, []byte) []byte) (int, error) {

	var lines, written int
	cap := int(buf.capacity)

	var b []byte

	var index int
	for i := offset; i < offset+rangeN; i++ {

		index = (cap - 1) - ((((-i - 1) + cap) % cap) % cap)

		if len(buf.data[index].Data) == 0 {
			continue
		}

		if ok := buf.filter.Test(buf.data[index].Label, buf.data[index].Data); !ok {
			continue
		}

		b = buf.data[index].Data
		for _, fn := range fns {
			b = fn(index, b)
		}

		// rangeN defines the max lines which can be currently displayed,
		// we need to take in account that lines in the buffer might wrap depending
		// on the current width of the screen resulting in writing >= rangeN lines
		// to the bytes.Buffer
		lines += bytes.Count(b, []byte("\n"))

		if lines >= rangeN {
			return written, nil
		}

		if _, err := w.Write(b); err != nil {
			return written, err
		}
		written++
	}
	return written, nil
}

var (
	ErrIndexOutOfBounds = fmt.Errorf("input index is grater than the capacity of the buffer or less than zero")
	ErrNotParsable      = fmt.Errorf("requested log line cannot be parsed to JSON")
	ErrMalformedLog     = fmt.Errorf("unable to format log. Log is malformed")
)

// At returns a single element in the buffer at the given index. It returns an error
// if the index is either grater than the buffer's capacity or less than 0.
// If the buffer has not overflown yet and the provided index is grater than the
// current buffer's write head-1 At still returns the value at the index however,
// it will be a nil byte slice.
// Since mutation of the item might occur the []byte slice is copied
func (buf Buffer) At(index int, fn func([]byte) ([]byte, error)) ([]byte, error) {
	if index > int(buf.capacity) || index < 0 {
		return nil, ErrIndexOutOfBounds
	}

	var item = make([]byte, len(buf.data[index].Data))
	copy(item, buf.data[index].Data)

	return fn(item)
}

//WithLineWrap wraps the slice of bytes based on the
// provided width where the resulting byte slice include
// \n after a maximum of width bytes.
func WithLineWrap(width int) func(int, []byte) []byte {
	return func(index int, b []byte) []byte {
		return wrap.Bytes(b, width)
	}
}

func WithInlineFormatting(width int, index int) func(int, []byte) []byte {
	return func(i int, b []byte) []byte {
		if i != index {
			return b
		}

		offset := bytes.IndexByte(b, byte('|'))

		data := b[offset+1:]
		// for some reason a lot of empty spaces are
		// added to the end of the styled string which
		// are messing up the formatting
		var cutoff = len(data) - 1
		for i := len(data) - 1; i >= 0; i-- {
			if data[i] != byte('\n') && data[i] != byte(' ') {
				break
			}
			cutoff = i
		}
		data = data[:cutoff]

		pretty, err := prettyjson.Format(data)
		if err != nil {
			return append(
				[]byte(
					lipgloss.NewStyle().
						Bold(true).
						Border(lipgloss.DoubleBorder(), false, false, false, true).
						BorderForeground(styles.DefaultColor.Border).
						Render(
							lipgloss.JoinVertical(lipgloss.Left,
								string(b[:offset]),
								string(wrap.Bytes(data[:cutoff], width-1)),
							),
						),
				),
				byte('\n'),
			)
		}
		return append(
			[]byte(
				lipgloss.NewStyle().
					Bold(true).
					Border(lipgloss.DoubleBorder(), false, false, false, true).
					BorderForeground(styles.DefaultColor.Border).
					Render(
						lipgloss.JoinVertical(lipgloss.Left,
							string(b[:offset]),
							string(wrap.Bytes(pretty, width-1)),
						),
					),
			),
			byte('\n'),
		)
	}
}
