package store

import (
	"fmt"
	"testing"
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
	}

	for _, tc := range tt {
		var depth int = 1
		lines := linewrap(&depth, tc.line, tc.width, 0)
		if depth != tc.depth {
			t.Fatalf("[%s] expected depth of: %d; got depth: %d and lines:\n\t%q", tc.name, tc.depth, depth, lines)
		}

		if lines != tc.expected {
			t.Fatalf("[%s] expected result of: %q; got result: %q", tc.name, tc.expected, lines)
		}
	}
}

func TestMoveDownNoOverflow(t *testing.T) {

	store := New(12)
	pager := store.NewPager(4, 10)

	sig := make(chan struct{})

	go func() {
		defer close(sig)

		for i := 0; i < 4; i++ {
			store.Insert("test-label", 0, []byte(fmt.Sprintf("Line-%d", i+1)))
			sig <- struct{}{}
		}
	}()

	sequence := []string{
		"Line-1",
		"Line-1\nLine-2",
		"Line-1\nLine-2\nLine-3",
		"Line-1\nLine-2\nLine-3\nLine-4",
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
	pager := store.NewPager(4, 10)

	sig := make(chan struct{})

	go func() {
		defer close(sig)

		for i := 0; i < 9; i++ {
			store.Insert("test-label", 0, []byte(fmt.Sprintf("Line-%d", i+1)))
			sig <- struct{}{}
		}
	}()

	sequence := []string{
		"Line-1",
		"Line-1\nLine-2",
		"Line-1\nLine-2\nLine-3",
		"Line-1\nLine-2\nLine-3\nLine-4",
		"Line-2\nLine-3\nLine-4\nLine-5",
		"Line-3\nLine-4\nLine-5\nLine-6",
		"Line-4\nLine-5\nLine-6\nLine-7",
		"Line-5\nLine-6\nLine-7\nLine-8",
		"Line-6\nLine-7\nLine-8\nLine-9",
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
