package ring

type Reader interface {
	At(i uint32) Item
	Range(start uint32, size uint8) []Item
}

type Item struct {
	Label       string
	Raw         string
	Parsed      string
	DataPointer int
	Revision    uint8
}

func (i Item) String() string {
	return i.Parsed
}

type Buffer struct {
	capacity uint32
	head     uint32
	data     []Item
}

func New(size uint32) Buffer {
	return Buffer{
		capacity: size,
		head:     0,
		data:     make([]Item, size),
	}
}

// Insert sets the given item at the next writing position
// of the buffer.
func (buf *Buffer) Insert(i Item) {
	buf.data[buf.head] = i
	buf.head = (buf.head + 1) % buf.capacity
}

// At returns an item at a given index of the buffer
func (buf *Buffer) At(i uint32) Item {
	return buf.data[buf.marshalIndex(i)]
}

func (buf *Buffer) Range(start uint32, size uint8) []Item {
	lower := buf.marshalIndex(start)
	upper := lower + uint32(size)

	return buf.data[lower:upper]
}

func (buf *Buffer) marshalIndex(absolute uint32) uint32 {
	return (buf.capacity - 1) - ((((-absolute - 1) + buf.capacity) % buf.capacity) % buf.capacity)
}
