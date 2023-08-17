package store

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

var (
	testRefreshRate = time.Nanosecond * 0 // evaluates to a nil time.Ticker causing asap update of pager
)

func TestMoveDownNoOverflow(t *testing.T) {

	store := New(12)
	pager := store.NewPager(4, 23, testRefreshRate)

	prefix := "test-label | "
	for i := 0; i < 4; i++ {
		store.Insert("test-label", len(prefix), []byte(fmt.Sprintf("%sLine-%d", prefix, i+1)))
	}

	sequence := []string{
		"test-label | Line-1\n\x00\n\x00\n\x00",
		"test-label | Line-1\ntest-label | Line-2\n\x00\n\x00",
		"test-label | Line-1\ntest-label | Line-2\ntest-label | Line-3\n\x00",
		"test-label | Line-1\ntest-label | Line-2\ntest-label | Line-3\ntest-label | Line-4",
	}

	// seqID := 0
	for _, seq := range sequence {
		pager.MoveDownDeprecated(false)
		if seq != pager.String() {
			t.Fatalf("[pager.MoveDown] expected line(s):\n%q\ngot:\n%q", seq, pager.String())
		}
	}
}

func TestMoveDownOverflow(t *testing.T) {

	store := New(12)
	pager := store.NewPager(4, 23, testRefreshRate)

	prefix := "test-label | "
	for i := 0; i < 9; i++ {
		store.Insert("test-label", len(prefix), []byte(fmt.Sprintf("%sLine-%d", prefix, i+1)))
	}

	sequence := []string{
		"test-label | Line-1\n\x00\n\x00\n\x00",
		"test-label | Line-1\ntest-label | Line-2\n\x00\n\x00",
		"test-label | Line-1\ntest-label | Line-2\ntest-label | Line-3\n\x00",
		"test-label | Line-1\ntest-label | Line-2\ntest-label | Line-3\ntest-label | Line-4",
		"test-label | Line-2\ntest-label | Line-3\ntest-label | Line-4\ntest-label | Line-5",
		"test-label | Line-3\ntest-label | Line-4\ntest-label | Line-5\ntest-label | Line-6",
		"test-label | Line-4\ntest-label | Line-5\ntest-label | Line-6\ntest-label | Line-7",
		"test-label | Line-5\ntest-label | Line-6\ntest-label | Line-7\ntest-label | Line-8",
		"test-label | Line-6\ntest-label | Line-7\ntest-label | Line-8\ntest-label | Line-9",
	}

	for _, seq := range sequence {
		pager.MoveDownDeprecated(false)
		if seq != pager.String() {
			t.Fatalf("[pager.MoveDown] expected line(s):\n%q\ngot:\n%q", seq, pager.String())
		}
	}
}

func TestMoveDownAssertHeight(t *testing.T) {

	// width := 20
	height := 9

	var store *Store
	var pager Pager

	prefix := "test-label | "
	tt := []struct {
		name      string
		maxWidth  int
		sequence  []string
		maxHeight int
		checksum  []string
	}{
		{
			name:      "each item fits in row ",
			maxHeight: 9,
			maxWidth:  35,
			sequence: []string{
				"test-label | Line-1",
				"test-label | Line-2",
				"test-label | Line-3",
				"test-label | Line-4",
				"test-label | Line-5",
				"test-label | Line-6",
				"test-label | Line-7",
				"test-label | Line-8",
				"test-label | Line-9",
			},
			checksum: []string{
				"test-label | Line-1",
				"test-label | Line-2",
				"test-label | Line-3",
				"test-label | Line-4",
				"test-label | Line-5",
				"test-label | Line-6",
				"test-label | Line-7",
				"test-label | Line-8",
				"test-label | Line-9",
			},
		},
		{
			name:      "overflow buffer; index prefix change",
			maxHeight: 9,
			maxWidth:  35,
			sequence: []string{
				"test-label | Line-1",
				"test-label | Line-2",
				"test-label | Line-3",
				"test-label | Line-4",
				"test-label | Line-5",
				"test-label | Line-6",
				"test-label | Line-7",
				"test-label | Line-8",
				"test-label | Line-9",
				"test-label | Line-10",
				"test-label | Line-11",
				"test-label | Line-12",
				"test-label | Line-13",
				"test-label | Line-14",
				"test-label | Line-15",
				"test-label | Line-16",
				"test-label | Line-17",
			},
			checksum: []string{
				"test-label | Line-9",
				"test-label | Line-10",
				"test-label | Line-11",
				"test-label | Line-12",
				"test-label | Line-13",
				"test-label | Line-14",
				"test-label | Line-15",
				"test-label | Line-16",
				"test-label | Line-17",
			},
		},
		{
			name:      "each item requires 2 lines",
			maxHeight: 9,
			maxWidth:  18,
			sequence: []string{
				"test-label | Line-10",
				"test-label | Line-20",
				"test-label | Line-30",
				"test-label | Line-40",
				"test-label | Line-50",
				"test-label | Line-60",
				"test-label | Line-70",
				"test-label | Line-80",
				"test-label | Line-90",
			},
			checksum: []string{
				"           | 50",
				"test-label | Line-",
				"           | 60",
				"test-label | Line-",
				"           | 70",
				"test-label | Line-",
				"           | 80",
				"test-label | Line-",
				"           | 90",
			},
		},
	}

	for _, tc := range tt {
		store = New(12)
		pager = store.NewPager(uint8(height), tc.maxWidth, testRefreshRate)

		for _, seq := range tc.sequence {
			store.Insert("test-label", len(prefix), []byte(seq))
			pager.MovePosition()
		}

		contents := pager.String()
		lineCount := strings.Count(contents, "\n")

		if lineCount > tc.maxHeight {
			t.Fatalf("[%s] wanted height: %d - got height: %d - content:\n%s", tc.name, tc.maxHeight, lineCount, contents)
		}

		testLines := strings.Split(contents, "\n")
		for i, line := range testLines {
			if tc.checksum[i] != line {
				t.Fatalf("[%s] content missmatch:\ncontent:\n%s\n\nwanted line: %s - got line: %s\n", tc.name, contents, tc.checksum[i], line)
			}
		}

	}
}

func TestAssertSameCapacity(t *testing.T) {

	capacity := 50
	store := New(1024)
	pager := store.NewPager(uint8(capacity), 100, testRefreshRate)

	prefix := "test-label | "
	for i := 0; i < 2048; i++ {
		store.Insert("test-label", len(prefix), []byte(`{"level":"error","ts":1692292122.983928,"caller":"application/structred.go:68","msg":"unable to do X","index":81,"error":"unable to do X","ts":1692292122.9839098,"stacktrace":"main.handleLog\n\t/Users/konstantingasser/coffecode/scotty/test/application/structred.go:68\nmain.main\n\t/Users/konstantingasser/coffecode/scotty/test/application/structred.go:47\nruntime.main\n\t/usr/local/go/src/runtime/proc.go:250"}`))
		pager.MovePosition()
		if cap(pager.buffer) != capacity {
			t.Fatalf("Capacity has changed after updating the buffer. From: %d -> To: %d", capacity, cap(pager.buffer))
		}
	}

}

// BenchmarkMoveDown-12        	16231768	       744.3 ns/op	    4033 B/op	       5 allocs/op
func BenchmarkMoveDown(b *testing.B) {
	store := New(2048)
	pager := store.NewPager(44, 75, testRefreshRate)

	// fill ring buffer until full so pager.position always is a hit
	for i := 0; i < 2048; i++ {
		store.Insert("dummy", len("dummy"), []byte(`{"level":"warn","ts":1680212791.946584,"caller":"application/structred.go:39","msg":"caution this indicates X","index":998,"ts":1680212791.946579}`))
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		pager.MoveDownDeprecated(false)
	}
}

// BenchmarkMovePosition-12    	54408824	       219.6 ns/op	     515 B/op	       4 allocs/op
func BenchmarkMovePosition(b *testing.B) {
	store := New(2048)
	pager := store.NewPager(44, 75, testRefreshRate)

	// fill ring buffer until full so pager.position always is a hit
	for i := 0; i < 2048; i++ {
		store.Insert("dummy", len("dummy"), []byte(`{"level":"warn","ts":1680212791.946584,"caller":"application/structred.go:39","msg":"caution this indicates X","index":998,"ts":1680212791.946579}`))
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		pager.MovePosition()
	}
}
