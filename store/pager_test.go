package store

import (
	"fmt"
	"testing"

	"github.com/KonstantinGasser/scotty/store/ring"
)

func TestLineWrap(t *testing.T) {

	tt := []struct {
		name     string
		line     string
		expected string
		width    int
		depth    int
	}{
		{
			name:     "wrap 3 times",
			line:     "foo bar baz",
			expected: "foo\n ba\nr b\naz",
			depth:    4,
			width:    3,
		},
		{
			name:     "line shorter than width",
			line:     "Hello, World!",
			width:    15,
			expected: "Hello, World!",
			depth:    1,
		},
		{
			name:     "line multiple times linger then width",
			line:     "Lorem ipsum dolor sit amet, consectetur adipiscing elit!",
			width:    10,
			expected: "Lorem ipsu\nm dolor si\nt amet, co\nnsectetur \nadipiscing\n elit!",
			depth:    6,
		},
		{
			name:     "line shorter than width",
			line:     "Line-1",
			width:    10,
			expected: "Line-1",
			depth:    1,
		},
		{
			name:     "width = len(line)",
			line:     "{'test': 'something'}",
			width:    21,
			expected: "{'test': 'something'}",
			depth:    1,
		},
	}

	for _, tc := range tt {
		// var depth int = 1
		depth, lines := linewrap(tc.line, tc.width, 0)

		if depth != tc.depth {
			t.Fatalf("[%s] expected depth of: %d; got depth: %d and lines:\n\t%q", tc.name, tc.depth, depth, lines)
		}

		if lines != tc.expected {
			t.Fatalf("[%s] expected result of: %q; got result: %q", tc.name, tc.expected, lines)
		}
	}
}

func TestBuildLine(t *testing.T) {

	tt := []struct {
		name       string
		item       ring.Item
		width      int
		wantHeight int
	}{
		{
			name: "unkown",
			item: ring.Item{
				Label:       "test label",
				DataPointer: len("test label") + 1,
				Raw:         "{'test': 'something'}",
			},
			width:      21,
			wantHeight: 2,
		},
	}

	for _, tc := range tt {
		_, _ = buildLine(tc.item, tc.width)
	}
}

func TestMoveDownNoOverflow(t *testing.T) {

	store := New(12)
	pager := store.NewPager(4, 20)

	sig := make(chan struct{})

	go func() {
		defer close(sig)

		prefix := "test-label | "
		for i := 0; i < 4; i++ {
			store.Insert("test-label", len(prefix), []byte(fmt.Sprintf("%sLine-%d", prefix, i+1)))
			sig <- struct{}{}
		}
	}()

	sequence := []string{
		"[1] test-label | Line-1",
		"[1] test-label | Line-1\n[2] test-label | Line-2",
		"[1] test-label | Line-1\n[2] test-label | Line-2\n[3] test-label | Line-3",
		"[1] test-label | Line-1\n[2] test-label | Line-2\n[3] test-label | Line-3\n[4] test-label | Line-4",
	}

	seqID := 0
	for range sig {
		pager.MoveDown()

		if sequence[seqID] != pager.String() {
			t.Fatalf("[pager.MoveDown] expected line(s):\n%q\ngot:\n%q", sequence[seqID], pager.String())
		}
		seqID++
	}
}

func TestMoveDownOverflow(t *testing.T) {

	store := New(12)
	pager := store.NewPager(4, 20)

	sig := make(chan struct{})

	go func() {
		defer close(sig)

		prefix := "test-label | "
		for i := 0; i < 4; i++ {
			store.Insert("test-label", len(prefix), []byte(fmt.Sprintf("%sLine-%d", prefix, i+1)))
			sig <- struct{}{}
		}
	}()

	sequence := []string{
		"[1] test-label | Line-1",
		"[1] test-label | Line-1\n[2] test-label | Line-2",
		"[1] test-label | Line-1\n[2] test-label | Line-2\n[3] test-label | Line-3",
		"[1] test-label | Line-1\n[2] test-label | Line-2\n[3] test-label | Line-3\n[4] test-label | Line-4",
		"[2] test-label | Line-2\n[3] test-label | Line-3\n[4] test-label | Line-4\n[5] test-label | Line-5",
		"[3] test-label | Line-3\n[4] test-label | Line-4\n[5] test-label | Line-5\n[6] test-label | Line-6",
		"[4] test-label | Line-4\n[5] test-label | Line-5\n[6] test-label | Line-6\n[7] test-label | Line-7",
		"[5] test-label | Line-5\n[6] test-label | Line-6\n[7] test-label | Line-7\n[8] test-label | Line-8",
		"[6] test-label | Line-6\n[7] test-label | Line-7\n[8] test-label | Line-8\n[9] test-label | Line-9",
	}

	seqID := 0
	for range sig {
		pager.MoveDown()

		if sequence[seqID] != pager.String() {
			t.Fatalf("[pager.MoveDown] expected line(s):\n%q\ngot:\n%q", sequence[seqID], pager.String())
		}
		seqID++
	}
}

func TestMoveDownAssertHeight(t *testing.T) {

	width := 18
	height := 4

	var store *Store
	var pager Pager
	// store := New(12)
	// pager := store.NewPager(uint8(height), width)

	prefix := "test-label | "
	tt := []struct {
		name      string
		sequence  []string
		maxHeight int
	}{
		{
			name:      "single item 1 lines < allowed height",
			maxHeight: 1,
			sequence: []string{
				"test-label | Line-1",
				// "test-label | Line-1\ntest-label | Line-2",
				// "test-label | Line-1\ntest-label | Line-2\ntest-label | Line-3",
				// "test-label | Line-1\ntest-label | Line-2\ntest-label | Line-3\ntest-label | Line-4",
				// "test-label | Line-2\ntest-label | Line-3\ntest-label | Line-4\ntest-label | Line-5",
				// "test-label | Line-3\ntest-label | Line-4\ntest-label | Line-5\ntest-label | Line-6",
				// "test-label | Line-4\ntest-label | Line-5\ntest-label | Line-6\ntest-label | Line-7",
				// "test-label | Line-5\ntest-label | Line-6\ntest-label | Line-7\ntest-label | Line-8",
				// "test-label | Line-6\ntest-label | Line-7\ntest-label | Line-8\ntest-label | Line-9",
			},
		},
	}
	for _, tc := range tt {
		store = New(12)
		pager = store.NewPager(uint8(height), width)

		for _, seq := range tc.sequence {
			store.Insert("test-label", len(prefix), []byte(seq))
		}

		pager.MoveDown()

		t.Log(pager.String())
	}
}

// func TestFormatFixedHeight(t *testing.T) {
// 	width := 21 // infers alls items must be broken into two lines
// 	height := 30

// 	store := New(50)
// 	pager := store.NewPager(uint8(height), width) // size of max 10 rows

// 	pager.EnableFormatting(2) // start does not really matter

// 	for i := 0; i < int(pager.size); i++ {
// 		pager.formatBuffer = append(pager.formatBuffer,
// 			ring.Item{
// 				Label:       "test label",
// 				DataPointer: len("test label") + 1,
// 				Raw:         "{'test': 'something'}",
// 			})
// 	}

// 	// pager.FormatNext() // should format

// 	contents := pager.String()
// 	got := strings.Count(contents, "\n")
// 	if got > height { // less is ok
// 		t.Fatalf("[format height check] wanted height: %d - got height: %d and content:\n\t%s\n", height, got, contents)
// 	}
// 	t.Log(contents)
// }

func BenchmarkMoveDown(b *testing.B) {
	store := New(2048)
	pager := store.NewPager(44, 100)

	// fill ring buffer until full so pager.position always is a hit
	for i := 0; i < 2048; i++ {
		store.Insert("dummy", 0, []byte(`{"level":"warn","ts":1680212791.946584,"caller":"application/structred.go:39","msg":"caution this indicates X","index":998,"ts":1680212791.946579}`))
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		pager.MoveDown()
	}
}
