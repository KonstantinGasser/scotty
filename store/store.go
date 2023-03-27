package store

import "github.com/KonstantinGasser/scotty/store/ring"

type Store struct {
	buffer ring.Buffer
}

func New(size uint32) *Store {
	return &Store{
		buffer: ring.New(size),
	}
}

func (store Store) NewPager(size uint8) Pager {

}
