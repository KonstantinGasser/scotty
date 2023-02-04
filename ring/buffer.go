package ring

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strconv"

	"github.com/KonstantinGasser/scotty/app/styles"
	"github.com/KonstantinGasser/scotty/debug"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wrap"
)

type Buffer struct {
	capacity uint32
	write    uint32
	data     [][]byte
}

// New initiates a new ring buffer with a set capacity.
// The provided factor can be any number however one must
// note that the capacity is the result of 1 << factor to
// ensure an end capacity of the power of 2.
// Example factors:
// 	factor of 10 -> 1 << 10 == 1024
// 	factor of 12 -> 1 << 12 == 4096
// The resulting capacity is the number of slots available in
// the ring buffer. To calculate the approximated memory size
// one has to take the size of the on average expected []byte
// stored in the buffer and compute: (1<<factor)*avg(item_size)
func New(factor uint32) Buffer {
	return Buffer{
		capacity: 1 << factor,
		write:    0,
		data:     make([][]byte, 1<<factor),
	}
}

func (buf Buffer) Cap() uint32 {
	return buf.capacity
}

func (buf *Buffer) Append(p []byte) {
	buf.data[buf.write] = p
	buf.write = (buf.write + 1) % buf.capacity
}

// Window write up to N of the last appended items to the io.Writer
// To modify items before writing them to the writer, a function can be provided.
//
func (buf Buffer) Window(w io.Writer, n int, fns ...func(int, []byte) []byte) error {

	// write := w.Write
	var writeIndex, cap int = int(buf.write), int(buf.capacity) // capture the latest write index
	var offset = writeIndex - n

	for i := offset; i < writeIndex; i++ {

		index := (cap - 1) - ((((-i - 1) + cap) % cap) % cap)

		val := buf.data[index]
		if val == nil {
			continue
		}

		for _, fn := range fns {
			val = fn(index, val)
		}

		// under the hood we pass in a bytes.Buffer
		// which again is using a slice of bytes where data
		// is appended to whenever write is called. However, this
		// is a potential bottleneck as runtime.growslice and
		// runtime.memmove will be called more frequently to adjust the
		// bytes.Buffer's buffer. Can be mitigated to a degree
		// by setting a capacity using Grow(N) where N is the educated guess
		// of how many bytes are expected to be written.
		if _, err := w.Write(val); err != nil {
			return err
		}
	}

	return nil
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

	var item = make([]byte, len(buf.data[index]))
	copy(item, buf.data[index])

	return fn(item)
}

func (buf Buffer) ScrollUp(w io.Writer, delta int, n int, fn func(int, []byte) []byte) error {

	var writeIndex, cap int = int(buf.write), int(buf.capacity)
	var offset = writeIndex - n - delta

	for i := offset; i < writeIndex-delta; i++ { // this loops over range [offset, writeIndex)

		index := (cap - 1) - ((((-i - 1) + cap) % cap) % cap)

		val := buf.data[index]
		if fn != nil {
			val = fn(index, val)
		}

		if _, err := w.Write(val); err != nil {
			return err
		}
	}

	return nil
}

// WithIndent returns a slice of byte in which
// the stream label and the formatted JSON are
// separated by the delimiter "@". If the data
// portion cannot be formatted the unformatted
// data is appended to the result slice.
func WithIndent() func([]byte) ([]byte, error) {
	return func(b []byte) ([]byte, error) {

		if b == nil {
			return nil, nil
		}

		offset := bytes.IndexByte(b, byte('|'))
		if offset < 0 {
			debug.Print("offset not found in -> %q\n", b)
			return nil, ErrMalformedLog
		}

		var label = b[0:offset]
		var data = bytes.TrimPrefix(
			bytes.TrimSuffix(
				b[offset+1:],
				[]byte("\n"),
			),
			[]byte(" "),
		)
		debug.Print("at offset: %q\n", data)

		var out bytes.Buffer
		if err := json.Indent(&out, data, " ", "\t"); err != nil {
			// in case of an error we don't care tbh. But at least show the
			// dev the log in unformatted
			return append(append(label, byte('@')), data...), nil
		}

		return append(append(label, byte('@')), out.Bytes()...), nil
	}
}

//WithLineWrap wraps the slice of bytes based on the
// provided width where the resulting byte slice include
// \n after a maximum of width bytes.
func WithLineWrap(width int) func(int, []byte) []byte {
	return func(index int, b []byte) []byte {
		return wrap.Bytes(append([]byte("["+strconv.Itoa(index)+"]"), b[:]...), width)
	}
}

func WithSelectedLine(index int) func(int, []byte) []byte {
	return func(i int, b []byte) []byte {
		if i != index {
			return b
		}

		offset := bytes.IndexByte(b, byte('|'))

		val := lipgloss.NewStyle().
			Foreground(styles.ColorBorder).
			Render(
				string(b[offset+1:]),
			)

		// for some reason a lot of empty spaces are
		// added to the end of the styled string which
		// are messing up the formatting
		var cutoff = len(val) - 1
		for i := len(val) - 1; i >= 0; i-- {
			if val[i] != byte('\n') && val[i] != byte(' ') {
				break
			}
			cutoff = i
		}

		return append(b[0:offset+1], val[:cutoff]...)
	}
}
