package store

import (
	"fmt"
	"strings"
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

func TestBreakLines(t *testing.T) {

	tt := []struct {
		name     string
		prefix   string
		line     string
		expected []string
		width    int
		depth    int
	}{
		{
			name:   "break 3 times",
			prefix: "",
			line:   "foo bar baz",
			expected: []string{
				"foo",
				" ba",
				"r b",
				"az",
			},
			depth: 4,
			width: 3,
		},
	}

	for _, tc := range tt {
		// var depth int = 1
		depth, lines := breaklines(tc.prefix, tc.line, tc.width, 0)

		if depth != tc.depth {
			t.Fatalf("[%s] expected depth of: %d; got depth: %d and lines:\n\t%q", tc.name, tc.depth, depth, lines)
		}

		for i, got := range lines {
			if got != tc.expected[i] {
				t.Fatalf("[%s] wanted: %s - got: %s", tc.name, strings.Join(tc.expected, ","), strings.Join(lines, ","))
			}
		}
	}
}

func TestBuildLines(t *testing.T) {

	tt := []struct {
		name       string
		item       ring.Item
		width      int
		wantHeight int
		wantLines  []string
	}{
		{
			name: "break line in two lines",
			item: ring.Item{
				Label:       "test label",
				DataPointer: len("test label") + 1,
				Raw:         "test label | {'test': 'something'}",
			},
			width:      15,
			wantHeight: 2,
			wantLines: []string{
				"[0] test label | {'test': 'som",
				"ething'}",
			},
		},
	}

	for _, tc := range tt {
		count, lines := buildLines(tc.item, tc.width+len(tc.item.Label)+4) // 4 -> preifx: "[x] " len=4

		if count != tc.wantHeight {
			t.Fatalf("[%s] wanted height: %d - got height: %d", tc.name, tc.wantHeight, count)
		}

		for i, line := range lines {
			if line != tc.wantLines[i] {
				t.Fatalf("[%s] wanted line: %s - got line: %s", tc.name, tc.wantLines[i], line)
			}
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

	prefix := "test-label | "
	for i := 0; i < 4; i++ {
		store.Insert("test-label", len(prefix), []byte(fmt.Sprintf("%sLine-%d", prefix, i+1)))
	}

	sequence := []string{
		"[1] test-label | Line-1\n\n\n",
		"[1] test-label | Line-1\n[2] test-label | Line-2\n\n",
		"[1] test-label | Line-1\n[2] test-label | Line-2\n[3] test-label | Line-3\n",
		"[1] test-label | Line-1\n[2] test-label | Line-2\n[3] test-label | Line-3\n[4] test-label | Line-4",
	}

	// seqID := 0
	for _, seq := range sequence {
		pager.MoveDown()
		fmt.Println(pager.String())
		fmt.Println("====")
		if seq != pager.String() {
			t.Fatalf("[pager.MoveDown] expected line(s):\n%q\ngot:\n%q", seq, pager.String())
		}
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
		"[1] test-label | Line-1\n\n\n\n\n\n\n",
		"[1] test-label | Line-1\n[2] test-label | Line-2\n\n\n\n\n\n",
		"[1] test-label | Line-1\n[2] test-label | Line-2\n[3] test-label | Line-3n\n\n\n\n",
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
			maxWidth:  20,
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
				"[1] test-label | Line-1",
				"[2] test-label | Line-2",
				"[3] test-label | Line-3",
				"[4] test-label | Line-4",
				"[5] test-label | Line-5",
				"[6] test-label | Line-6",
				"[7] test-label | Line-7",
				"[8] test-label | Line-8",
				"[9] test-label | Line-9",
			},
		},
		// {
		// 	name:      "overflow buffer; index prefix change",
		// 	maxHeight: 9,
		// 	maxWidth:  22,
		// 	sequence: []string{
		// 		"test-label | Line-1",
		// 		"test-label | Line-2",
		// 		"test-label | Line-3",
		// 		"test-label | Line-4",
		// 		"test-label | Line-5",
		// 		"test-label | Line-6",
		// 		"test-label | Line-7",
		// 		"test-label | Line-8",
		// 		"test-label | Line-9",
		// 		"test-label | Line-10",
		// 		"test-label | Line-11",
		// 		"test-label | Line-12",
		// 		"test-label | Line-13",
		// 		"test-label | Line-14",
		// 		"test-label | Line-15",
		// 		"test-label | Line-16",
		// 		"test-label | Line-17",
		// 	},
		// 	checksum: []string{
		// 		"[9] test-label | Line-9",
		// 		"[10] test-label | Line-10",
		// 		"[11] test-label | Line-11",
		// 		"[12] test-label | Line-12",
		// 		"[1] test-label | Line-13",
		// 		"[2] test-label | Line-14",
		// 		"[3] test-label | Line-15",
		// 		"[4] test-label | Line-16",
		// 		"[5] test-label | Line-17",
		// 	},
		// },
		// {
		// 	name:      "each item requires 2 lines",
		// 	maxHeight: 9,
		// 	maxWidth:  18,
		// 	sequence: []string{
		// 		"test-label | Line-1",
		// 		"test-label | Line-2",
		// 		"test-label | Line-3",
		// 		"test-label | Line-4",
		// 		"test-label | Line-5",
		// 		"test-label | Line-6",
		// 		"test-label | Line-7",
		// 		"test-label | Line-8",
		// 		"test-label | Line-9",
		// 	},
		// 	checksum: []string{
		// 		"-5",
		// 		"[6] test-label | Line",
		// 		"-6",
		// 		"[7] test-label | Line",
		// 		"-7",
		// 		"[8] test-label | Line",
		// 		"-8",
		// 		"[9] test-label | Line",
		// 		"-9",
		// 	},
		// },
	}
	for _, tc := range tt {
		store = New(12)
		pager = store.NewPager(uint8(height), tc.maxWidth)

		for _, seq := range tc.sequence {
			store.Insert("test-label", len(prefix), []byte(seq))
			pager.MoveDown()
		}

		contents := pager.String()
		lineCount := strings.Count(contents, "\n")

		if lineCount > tc.maxHeight {
			t.Fatalf("[%s] wanted height: %d - got height: %d - content:\n%s", tc.name, tc.maxHeight, lineCount, contents)
		}

		for i, line := range strings.Split(contents, "\n") {
			if tc.checksum[i] != line {
				t.Fatalf("[%s] content missmatch:\ncontent:\n%s\n\nwanted line: %s - got line: %s\n", tc.name, contents, tc.checksum[i], line)
			}
		}

	}
}

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
