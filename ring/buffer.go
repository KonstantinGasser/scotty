package ring

import (
	"io"
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

func (buf *Buffer) Append(p []byte) {
	buf.data[buf.write] = p
	buf.write = (buf.write + 1) % buf.capacity
}

// Window write up to N of the last appended items to the io.Writer
// To modify items before writing them to the writer, a function can be provided.
//
func (buf Buffer) Window(w io.Writer, n int, fn func([]byte) []byte) error {

	// write := w.Write
	var writeIndex, cap int = int(buf.write), int(buf.capacity) // capture the latest write index
	var offset = writeIndex - n

	for i := offset; i < writeIndex; i++ {

		index := (cap - 1) - ((((-i - 1) + cap) % cap) % cap)

		val := buf.data[index]
		if val == nil {
			continue
		}

		if fn != nil {
			val = fn(val)
		}

		// under the hood we pass in a strings.Builder/bytes.Buffer
		// which again is using a slice of bytes where data
		// is appended to whenever write is called. However, this
		// is a potential bottleneck as runtime.growslice and
		// runtime.memmove will be called more frequently to adjust the
		// strings.Builder/bytes.Buffer buffer. Can be mitigated somehow
		// by setting a capacity using Grow(N) where N is the educated guess
		// of how many bytes are expected to be written.
		if _, err := w.Write(val); err != nil {
			return err
		}
	}

	return nil
}

func (buf Buffer) ScrollUp(w io.Writer, delta int, n int, fn func([]byte) []byte) error {

	var writeIndex, cap int = int(buf.write), int(buf.capacity)
	var offset = writeIndex - n - delta

	for i := offset; i < writeIndex-delta; i++ { // this loops over range [offset, writeIndex)

		index := (cap - 1) - ((((-i - 1) + cap) % cap) % cap)

		val := buf.data[index]
		if fn != nil {
			val = fn(val)
		}

		if _, err := w.Write(val); err != nil {
			return err
		}
	}

	return nil
}
