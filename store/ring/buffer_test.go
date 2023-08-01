package ring

import (
	"fmt"
	"testing"
)

func TestRangeNoOverflow(t *testing.T) {

	buffer := New(12)

	for i := 0; i < 12; i++ {
		buffer.Insert(Item{
			Raw: fmt.Sprintf("Line-%d", i+1),
		})
	}

	tt := []struct {
		name  string
		start int
		size  int
		want  []Item
	}{
		{
			name:  "range all",
			start: 0,
			size:  12,
			want: []Item{
				{Raw: "Line-1"},
				{Raw: "Line-2"},
				{Raw: "Line-3"},
				{Raw: "Line-4"},
				{Raw: "Line-5"},
				{Raw: "Line-6"},
				{Raw: "Line-7"},
				{Raw: "Line-8"},
				{Raw: "Line-9"},
				{Raw: "Line-10"},
				{Raw: "Line-11"},
				{Raw: "Line-12"},
			},
		},
		{
			name:  "range middel part",
			start: 4,
			size:  8,
			want: []Item{
				{Raw: "Line-5"},
				{Raw: "Line-6"},
				{Raw: "Line-7"},
				{Raw: "Line-8"},
				{Raw: "Line-9"},
			},
		},
		{
			name:  "random range",
			start: 5,
			size:  8,
			want: []Item{
				{Raw: "Line-6"},
				{Raw: "Line-7"},
				{Raw: "Line-8"},
				{Raw: "Line-9"},
				{Raw: "Line-10"},
				{Raw: "Line-11"},
				{Raw: "Line-12"},
				{Raw: "Line-1"},
			},
		},
	}

	for _, tc := range tt {
		res := buffer.Range(tc.start, tc.size)

		for i, item := range tc.want {
			if item.Raw != res[i].Raw {
				t.Fatalf("[%s] wanted item: %s; got item: %s", tc.name, item.Raw, res[i].Raw)
			}
		}
	}
}

func TestRangeOverflow(t *testing.T) {

	buffer := New(12)

	for i := 0; i < 18; i++ { // overflow by 18-12 => 6 items
		buffer.Insert(Item{
			Raw: fmt.Sprintf("Line-%d", i+1),
		})
	}

	tt := []struct {
		name  string
		start int
		size  int
		want  []Item
	}{
		{
			name:  "range with no overflow part",
			start: 6,
			size:  10,
			want: []Item{
				{Raw: "Line-7"},
				{Raw: "Line-8"},
				{Raw: "Line-9"},
				{Raw: "Line-10"},
			},
		},
		{
			name:  "range with overflow part",
			start: 10,
			size:  4,
			want: []Item{
				{Raw: "Line-11"},
				{Raw: "Line-12"},
				{Raw: "Line-13"},
				{Raw: "Line-14"},
			},
		},
	}

	for _, tc := range tt {
		res := buffer.Range(tc.start, tc.size)

		for i, item := range tc.want {
			if item.Raw != res[i].Raw {
				t.Fatalf("[%s] wanted item: %s; got item: %s", tc.name, item.Raw, res[i].Raw)
			}
		}
	}
}

func TestRangeWithEmptyValues(t *testing.T) {

	buffer := New(12)

	for i := 0; i < 6; i++ { // overflow by 18-12 => 6 items
		buffer.Insert(Item{
			Raw: fmt.Sprintf("Line-%d", i+1),
		})
	}

	tt := []struct {
		name  string
		start int
		size  int
		want  []Item
	}{
		{
			name:  "range all",
			start: 0,
			size:  12,
			want: []Item{
				{Raw: "Line-1"},
				{Raw: "Line-2"},
				{Raw: "Line-3"},
				{Raw: "Line-4"},
				{Raw: "Line-5"},
				{Raw: "Line-6"},
				{Raw: ""},
				{Raw: ""},
				{Raw: ""},
				{Raw: ""},
				{Raw: ""},
				{Raw: ""},
			},
		},
		{
			name:  "range middel part",
			start: 4,
			size:  8,
			want: []Item{
				{Raw: "Line-5"},
				{Raw: "Line-6"},
				{Raw: ""},
				{Raw: ""},
				{Raw: ""},
			},
		},
	}

	for _, tc := range tt {
		res := buffer.Range(tc.start, tc.size)

		for i, item := range tc.want {
			if item.Raw != res[i].Raw {
				t.Fatalf("[%s] wanted item: %s; got item: %s", tc.name, item.Raw, res[i].Raw)
			}
		}
	}
}

func newFilledBuffer(size int, writes int, fn func(i int) string) *Buffer {

	buf := New(uint32(size))
	for i := 0; i < writes; i++ {
		buf.Insert(Item{
			Raw: fn(i),
		})
	}
	return buf
}

func TestOffsetRead(t *testing.T) {

	tt := []struct {
		name     string
		bufSize  int
		bufWrite int
		offset   int
		pageSize int
		want     []Item
	}{
		{
			name:     "offset start zero",
			bufSize:  32,
			bufWrite: 32,
			offset:   0,
			pageSize: 16,
			want: []Item{
				{
					Raw: "Line-0",
				},
				{
					Raw: "Line-1",
				},
				{
					Raw: "Line-2",
				},
				{
					Raw: "Line-3",
				},
				{
					Raw: "Line-4",
				},
				{
					Raw: "Line-5",
				},
				{
					Raw: "Line-6",
				},
				{
					Raw: "Line-7",
				},
				{
					Raw: "Line-8",
				},
				{
					Raw: "Line-9",
				},
				{
					Raw: "Line-10",
				},
				{
					Raw: "Line-11",
				},
				{
					Raw: "Line-12",
				},
				{
					Raw: "Line-13",
				},
				{
					Raw: "Line-14",
				},
				{
					Raw: "Line-15",
				},
			},
		},
		{
			name:     "offset; buffer wrapped",
			bufSize:  32,
			bufWrite: 44,
			offset:   12,
			pageSize: 6,
			want: []Item{
				{
					Raw: "Line-12",
				},
				{
					Raw: "Line-13",
				},
				{
					Raw: "Line-14",
				},
				{
					Raw: "Line-15",
				},
				{
					Raw: "Line-16",
				},
				{
					Raw: "Line-17",
				},
			},
		},
	}
	for _, tc := range tt {

		buffer := newFilledBuffer(tc.bufSize, tc.bufWrite, func(i int) string { return fmt.Sprintf("Line-%d", i) })

		got := make([]Item, tc.pageSize)
		buffer.OffsetRead(tc.offset, got)

		if len(got) != len(tc.want) {
			t.Fatalf("[%s -- offset: %d] wanted slice of len: %d, got len: %d", tc.name, tc.offset, len(tc.want), len(got))
		}
		for i, item := range tc.want {
			if got[i].Raw != item.Raw {
				t.Fatalf("[%s -- offset: %d] want: %q, got: %q", tc.name, tc.offset, item.Raw, got[i].Raw)
			}
		}
	}
}

/*
Current benchmark results

goos: darwin
goarch: arm64
pkg: github.com/KonstantinGasser/scotty/store/ring
BenchmarkOffsetRead
BenchmarkOffsetRead-12    	10446462	        97.18 ns/op	       0 B/op	       0 allocs/op
*/
func BenchmarkOffsetRead(b *testing.B) {

	buf := New(2048)

	for i := 0; i < 2048; i++ {
		buf.Insert(Item{
			Label:       "hello-world",
			Raw:         `hello-world | {"level":"warn","ts":1680212791.946584,"caller":"application/structred.go:39","msg":"caution this indicates X","index":998,"ts":1680212791.946579}`,
			DataPointer: 14,
		})
	}

	avgRange := 50

	items := make([]Item, avgRange)

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		buf.OffsetRead(avgRange, items)
	}

}
