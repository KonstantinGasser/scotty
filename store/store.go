package store

import (
	"sync"

	"github.com/KonstantinGasser/scotty/store/ring"
)

/*

REQUIREMENTS FOR THE STORE

1. The store needs to be able to insert data quickly -> writes over reads
1.1 At least needs to be quick such that the new log is quickly available to be displayed in the tail-view
1.2 Indexing, and inserting the data the table can be deferred in a background job/worker

2. The store must allow to retrieve X records (order by their time of write) where X is defined by the range [N,M) with the option to apply filter (FilterByKey("level", "error"), etc.)
2.1 If no filters are set this operation should be achieved in O(1) time complexity (logs[N:M])
2.2 If filters are set the operation might go up to O(n) time complexity

3. The store (or table) should allow to set indices to allow for a O(log n) time complexity for queries

4. All data should be stored in-memory and does not need to be persisted between restarts of scotty

5. Also it would be nice to limit the memory consumption


QUESTIONS

1. When a stream disconnects it can be assumed that once it's connecting again a new version of the app is streaming data (developer made changes to the application, etc.).
Should logs prior to the new "version" be deleted? On the one hand, yes - because queries should ignore logs from an "dead" version, on the other hand, no because the developer
might want to scroll up viewing the logs of the prior version.

*/

// Log represents a single log stored for tailing
// a window of logs including the stream from which
// the log was emitted
type Log struct {
	Stream string
	Data   []byte
}

type Store struct {
	// guards the index log and tables map
	mtx sync.RWMutex

	// tail holds up to max_buf_size of the latest logs and is
	// an immutable append-only data structure.
	// Due to the nature of a ring buffer once the buffer
	// is full starting from the oldest entry entries will
	// be dropped. This however, is ok since we use this tail
	// buffer to scroll/brows through the logs (pager.Logger).
	// The user itself can increase the buffer size if needed,
	// default are 4096 entries
	// Furthermore, we do not need to be concerned about deleting
	// entries if a stream disconnects as the user might want to read
	// through the disconnected stream logs.
	tail *ring.Buffer[Log]

	// tables hold stream specific logs. Each log must belong
	// to on table. A table must exists when the first log is
	// received - table creation should happen when a beam syncs
	// scotty
	// tables map[string]Table
}

// Insert appends the given value to the store.
// Depending on the label of the value (refers to the table)
// the value is forwarded to the respective table to handle the insert.
// All present fields within value are added to the tables bloom-filter
func (s *Store) Insert(v []byte) {

}

// Window returns all captured logs within the passed
// boundaries [top, bottom). If top >= bottom or top exceeds
// the buffer cap or the bottom < 0 Window returns an empty slice.
func (s *Store) Window(top int, bottom int) []Log {
	return s.tail.Seek(top, bottom)
}

func (s *Store) Query() *Query {
	return nil
}
