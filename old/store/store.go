package store

type Store struct {
	buffer *ringbuf
}

func New(capacity uint16) *Store {
	return &Store{}
}

func (s Store) Add(value interface{}) {
	s.buffer.queue(value)
}

func (s Store) Tail(n int) []interface{} {
	return s.buffer.tailN(n)
}
