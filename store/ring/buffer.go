package ring

type Reader interface {
	At(i uint32) Item
	Range(start int, size int) []Item
}

type Item struct {
	index       uint32
	Label       string
	Raw         string
	Parsed      string
	DataPointer int
	Revision    uint8
}

func (i Item) String() string {
	return i.Parsed
}

func (i Item) Index() uint32 {
	return i.index
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
	i.index = buf.head + 1 // plus one since eventhough we start at zero the first log should show 1 not 0 as index
	buf.data[buf.head] = i
	buf.head = (buf.head + 1) % buf.capacity
}

// At returns an item at a given index of the buffer
func (buf *Buffer) At(i uint32) Item {
	return buf.data[buf.marshalIndex(i)]
}

// Range returns a slice starting somewhere in the buffer
// with all items sequentally till the given size.
//
// Range does not care about dirty-reads (not the ACID dirty reads)
// but rather Range does not check if the requested range is crossing
// the end of the buffer resulting in the latest items of the buffer
// at the beginning of the returned slice while the next items are the
// oldest items in the buffer
func (buf *Buffer) Range(start int, size int) []Item {

	var out []Item = make([]Item, 0, size)
	var index uint32

	for i := start; i < start+size; i++ {
		index = buf.marshalIndex(uint32(i))

		if len(buf.data[index].Raw) <= 0 {
			continue
		}

		out = append(out, buf.data[index])
	}

	return out
}

func (buf *Buffer) marshalIndex(absolute uint32) uint32 {
	return ((absolute % buf.capacity) + buf.capacity) % buf.capacity
}
