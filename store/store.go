package store

import (
	"sync"
	"unsafe"
)

type Store struct {
	// guards the index log and tables map
	mtx sync.RWMutex
	// index is an append-only immutable slice
	// to which each log to any table is stored
	// in the order received by scotty.
	// This index allows to yield back logs from
	// N tables of an range [x,z).
	// The index does not store the actual data but
	// rather the reference (memory address).
	// A drawback to overcome are tombstone elements
	// which are a result of tables being dropped after
	// a stream disconnects. In the future we should
	// implement a background task which removes tombstone
	// elements from the index effectively doing a indexing
	// by removing nil pointer.
	// Keeping track of the index is required to display
	// N logs (where N most likely corresponds to viewport.Height)
	// starting from [x,z). This allows to an O(1) constant time
	// to retrieve N logs rather then mashing the tables together
	// and sorting them by time-of-arrival
	index []unsafe.Pointer

	// tables hold stream specific logs. Each log must belong
	// to on table. A table must exists when the first log is
	// received - table creation should happen when a beam syncs
	// scotty
	tables map[string]Table
}

// Insert appends the given value to the store.
// Depending on the label of the value (refers to the table)
// the value is forwarded to the respective table to handle the insert.
// All present fields within value are added to the tables bloom-filter
func (s *Store) Insert(v map[string]interface{}) {

}

// Slice returns N values
// Tombstones within the index of the store are ignored and
// are not counted towards the range [start, end).
//
// QUESTION @KonstantinGasser:
// checking for tombstones results in an O(n) time complexity,
// is that something we can neglect since the delta from start to end
// most likely is never higher then the height of the terminal?
// Furthermore, we should think about what we need. The viewport
// which most likely is calling the API would need to iterate again
// and construct the string to display in the viewport. As such we could
// offer an API called Tail(start int, end int) string which returns
// the constructed string?
func (s *Store) Slice(start int, end int) []map[string]interface{} {
	return nil
}

func (s *Store) Query() *Query {
	return nil
}
