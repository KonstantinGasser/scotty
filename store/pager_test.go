package store

import (
	"fmt"
	"strings"
	"testing"
)

func TestMoveDownNoOverflow(t *testing.T) {

	store := New(12)
	pager := store.NewPager(4, 23)

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
		if seq != pager.String() {
			t.Fatalf("[pager.MoveDown] expected line(s):\n%q\ngot:\n%q", seq, pager.String())
		}
	}
}

func TestMoveDownOverflow(t *testing.T) {

	store := New(12)
	pager := store.NewPager(4, 23)

	prefix := "test-label | "
	for i := 0; i < 9; i++ {
		store.Insert("test-label", len(prefix), []byte(fmt.Sprintf("%sLine-%d", prefix, i+1)))
	}

	sequence := []string{
		"[1] test-label | Line-1\n\n\n",
		"[1] test-label | Line-1\n[2] test-label | Line-2\n\n",
		"[1] test-label | Line-1\n[2] test-label | Line-2\n[3] test-label | Line-3\n",
		"[1] test-label | Line-1\n[2] test-label | Line-2\n[3] test-label | Line-3\n[4] test-label | Line-4",
		"[2] test-label | Line-2\n[3] test-label | Line-3\n[4] test-label | Line-4\n[5] test-label | Line-5",
		"[3] test-label | Line-3\n[4] test-label | Line-4\n[5] test-label | Line-5\n[6] test-label | Line-6",
		"[4] test-label | Line-4\n[5] test-label | Line-5\n[6] test-label | Line-6\n[7] test-label | Line-7",
		"[5] test-label | Line-5\n[6] test-label | Line-6\n[7] test-label | Line-7\n[8] test-label | Line-8",
		"[6] test-label | Line-6\n[7] test-label | Line-7\n[8] test-label | Line-8\n[9] test-label | Line-9",
	}

	for _, seq := range sequence {
		pager.MoveDown()
		if seq != pager.String() {
			t.Fatalf("[pager.MoveDown] expected line(s):\n%q\ngot:\n%q", seq, pager.String())
		}
	}
}

// func TestMoveDownExeedPagerBufferSize(t *testing.T) {

// 	store := New(12)
// 	pager := store.NewPager(4, 20)

// 	prefix := "test-label | "
// 	for i := 0; i < 2; i++ {
// 		store.Insert("test-label", len(prefix), []byte(fmt.Sprintf("%sLine-%d", prefix, i+1)))
// 	}

// 	// insert log line which if broken into multiple lines
// 	// will have more lines then the pagers height allows
// 	// this item if put set into view by the MoveDown call
// 	// require 8 lines of space - the pagegr however is only
// 	// 4 lines heigh
// 	store.Insert("test-lable", len(prefix), []byte("test-labe2 | aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"))

// 	sequence := []string{
// 		"[1] test-label | Line-1\n\n\n",
// 		"[1] test-label | Line-1\n[2] test-label | Line-2\n\n",
// 		"[1] test-label | Line-1\n[2] test-label | Line-2\n",
// 		"[1] test-label | Line-1\n[2] test-label | Line-2\n[3] test-label | Line-3\n[4] test-label | Line-4",
// 	}

// 	for range sequence {
// 		pager.MoveDown()
// 		t.Log("===========")
// 		t.Log(pager.String())
// 		// if seq != pager.String() {
// 		// 	t.Fatalf("[pager.MoveDown] expected line(s):\n%q\ngot:\n%q", seq, pager.String())
// 		// }
// 	}
// }

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
			maxWidth:  30,
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
		{
			name:      "overflow buffer; index prefix change",
			maxHeight: 9,
			maxWidth:  30,
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
				"[9] test-label | Line-9",
				"[10] test-label | Line-10",
				"[11] test-label | Line-11",
				"[12] test-label | Line-12",
				"[13] test-label | Line-13",
				"[14] test-label | Line-14",
				"[15] test-label | Line-15",
				"[16] test-label | Line-16",
				"[17] test-label | Line-17",
			},
		},
		{
			name:      "each item requires 2 lines",
			maxHeight: 9,
			maxWidth:  18,
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
				"ine-5",
				"[6] test-label | L",
				"ine-6",
				"[7] test-label | L",
				"ine-7",
				"[8] test-label | L",
				"ine-8",
				"[9] test-label | L",
				"ine-9",
			},
		},
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
