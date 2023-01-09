package store

import "github.com/bits-and-blooms/bloom/v3"

type Entry struct {
	Key   string
	Value any
}

type Table struct {
	// identifies a table with a unique name such as
	// the stream label. If a Table.Identifier is
	// already present the table must not be created
	identifier string
	// catalog keeps an index of columns/fields
	// present in the table with at least one entry.
	// Due to the nature of bloom filter the answer to
	// the asked question might be false positive.
	catalog bloom.BloomFilter
	// entries is a slice implemented as a ring-buffer
	// for a fixed memory allocation where each entry
	// represents a single log parsed to a structured form
	entires []Entry
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

TL;CT
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
