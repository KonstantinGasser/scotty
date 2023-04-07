package store

import (
	"github.com/KonstantinGasser/scotty/debug"
	"github.com/KonstantinGasser/scotty/store/ring"
)

type Store struct {
	buffer ring.Buffer
}

func New(size uint32) *Store {
	return &Store{
		buffer: ring.New(size),
	}
}

func (store *Store) Insert(label string, offset int, data []byte) {
	store.buffer.Insert(ring.Item{
		Label:       label,
		Raw:         string(data),
		DataPointer: offset,
	})
}

func (store Store) NewPager(size uint8, width int) Pager {

	var buffer = make([]string, size)
	for i := range buffer {
		buffer[i] = " "
	}
	pager := Pager{
		size:       size,
		ttyWidth:   width,
		reader:     &store.buffer,
		position:   0,
		mode:       tailing,
		pageOffset: 0,
		buffer:     buffer,
		written:    0,
		raw:        "",
	}

	debug.Print("pager.buffer => len: %d\n", len(pager.buffer))
	return pager
}
