package store

import (
	"sync"
	"unsafe"

	"github.com/bits-and-blooms/bloom/v3"
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
func (s *Store) Slice(start int, end int) []map[string]interface{} {
	return nil
}

type Query struct{}

func (s *Store) Query() *Query {
	return nil
}

func (q *Query) Exec() error {
	return nil
}

type Entry struct {
	Key   string
	Value any
}

type Table struct {
	// Identifies a table with a unique name such as
	// the stream label. If a Table.Identifier is
	// already present the table must not be created
	Identifier string
	// Catalog keeps an index of columns/fields
	// present in the table with at least one entry.
	// Due to the nature of bloom filter the answer to
	// the asked question might be false positive.
	Catalog bloom.BloomFilter
	// Entries is a slice implemented as a ring-buffer
	// for a fixed memory allocation where each entry
	// represents a single log parsed to a structured form
	Entires []Entry
}

/*

BIG QUESTION: @KonstantinGasser:
so now each stream has a stable in which the respective
logs are stored; great. However, what's your plan on retrieving N elements
across all tables? The question to answer is with the data separated by streams
we need to looks up N data from M tables where the 1..N must be in sequential
order the logs where received by scotty.
The store somehow needs to keep an list of the ordered values.

Follow-up question once a stream disconnects we want to be able to remove all current
values belonging to the stream which means the store needs an somewhat good and performant
way to delete all of the these values. Using an array (ordered by the ts arrival at scotty)
we will have O(n) - keep in mind it might happen regularly but when it happens time complexity
is not killing us

TLCT ()
What about an append-log type of thing?
Say each log which is streamed at added to the append-log,
reads are done from the back. each entry keeps a pointer to a specific log
in memory and if we drop a table gc will take care of the rest?? Ok sure we cannot
trust gc in a time sense means a nil check of the pointer is required

Using tables per stream

Table: app_1

Entries (could still be a ring buffer to limit the amount of memory):
| level | error | any_1 | any_2 | ... | any_n |
-----------------------------------------------

Catalog: app_1
The catalog of keeps track of columns which are available in table.
This is due to the fact that we do not know the schema beforehand
and the schema is gradually extended when JSON fields not present yet
in the table are added. Having the catalog allows queries to determine
if it is worth scanning a particular table or not

{
	level : ok
	error : !ok
	any_1 : !ok
	any_2 : ok
}

If no indices are provided try to guess some??
-- educated guess could be "level", "error"?

*/
