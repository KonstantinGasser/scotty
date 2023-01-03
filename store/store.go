package store

import "github.com/bits-and-blooms/bloom/v3"

type Store struct{}

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

*/
