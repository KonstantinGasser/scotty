package ring

import (
	"io"
)

type Buffer struct {
	capacity uint32
	write    uint32
	read     uint32
	data     [][]byte
}

// New initiates are new ring buffer with a set capacity.
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
func New(factor uint32) *Buffer {
	return &Buffer{
		capacity: 1 << factor,
		write:    0,
		read:     0,
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
func (buf Buffer) Window(w io.Writer, n uint32, fn func([]byte) []byte) error {

	var window int

	if n < buf.read {
		window = int(buf.read - n)
	}

	for i := window; i < int(buf.read); i++ {
		w.Write(buf.data[i])
	}

	return nil
}
