package store

import (
	"strings"
	"time"

	"github.com/KonstantinGasser/scotty/store/ring"
)

type Store struct {
	buffer *ring.Buffer
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

func (store Store) NewPager(size uint8, width int, refresh time.Duration) Pager {
	buf := make([]string, size)
	for i := range buf {
		buf[i] = "\000"
	}

	var ticker *time.Ticker
	if refresh > 0 {
		ticker = time.NewTicker(refresh)
	}
	return Pager{
		size:       size,
		ttyWidth:   width,
		reader:     store.buffer,
		position:   0,
		buffer:     buf,
		written:    0,
		bufferView: strings.Join(buf, "\n"),
		ticker:     ticker,
	}
}

func (store Store) NewFormatter(size uint8, width int) Formatter {
	return Formatter{
		size:     size,
		ttyWidth: width,
		reader:   store.buffer,
		absolute: 0,
		relative: 0,
	}
}
