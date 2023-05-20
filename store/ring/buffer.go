package ring

type Reader interface {
	At(i uint32) Item
	Range(start int, size int) Slice
	Head() uint32
}

type Slice []Item

// Strings transforms the slice of Items to a slice
// of strings where the call determins which struct field
// should be included in the resulting slice.
func (s Slice) Strings(fn func(i Item) string) []string {
	var out = make([]string, len(s))

	for i, item := range s {
		out[i] = fn(item)
	}

	return out
}

// Item represents one element in the Buffer.
// While fields such as the index or the label
// seem clear the field Raw needs some explaining.
// The fiel Raw includes the entire row/line finished
// formatted/colored/build stored at a given index.
// In order to retrieve only the application log
// Item has a DataPointer which must be provided and
// indicated at which index of Raw the application log
// start. To get only the application log one can do
// the following:
//
//	item = Item {
//			index: 0,
//			Label: "test"
//			Raw: "test | level=debug, msg=I am the application log",
//			DataPointer: 7, // -> len("test | ") where "test | " is the formatted/build line prefix
//	}
//
// log := item.Raw[item.DataPointer:] // == log := "level=debug, msg=I am the application log"
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
	written  uint32
	data     []Item
}

func New(size uint32) Buffer {
	return Buffer{
		capacity: size,
		head:     0,
		written:  0,
		data:     make([]Item, size),
	}
}

// Insert sets the given item at the next writing position
// of the buffer.
func (buf *Buffer) Insert(i Item) {
	buf.written += 1
	i.index = buf.written

	buf.data[buf.head] = i
	buf.head = (buf.head + 1) % buf.capacity
}

// At returns an item at a given index of the buffer
func (buf *Buffer) At(i uint32) Item {
	return buf.data[buf.marshalIndex(i)]
}

// Head returns the latest index written to
func (buf Buffer) Head() uint32 {
	return buf.head
}

// Range returns a slice starting somewhere in the buffer
// with all items sequentally till the given size.
//
// Range does not care about dirty-reads (not the ACID dirty reads)
// but rather Range does not check if the requested range is crossing
// the end of the buffer.head resulting in the latest items of the buffer
// at the beginning of the returned slice while the next items are the
// oldest items in the buffer.
// Range add items regardless of there zero value.
// idea: func(start, buf []Item) size based on len(buf)?
func (buf *Buffer) Range(start int, size int) Slice {

	var out []Item = make([]Item, 0, size)
	var index uint32

	for i := start; i < start+size; i++ {
		index = buf.marshalIndex(uint32(i))

		out = append(out, buf.data[index])
	}

	return out
}

func (buf *Buffer) marshalIndex(absolute uint32) uint32 {
	return ((absolute % buf.capacity) + buf.capacity) % buf.capacity
}
