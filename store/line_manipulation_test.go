package store

import (
	"strings"
	"testing"

	"github.com/KonstantinGasser/scotty/store/ring"
)

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
		depth, lines := breaklines(tc.prefix, 0, tc.line, tc.width, 0)

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
				DataPointer: len("test label | "),
				Raw:         "test label | {'test': 'something'}",
			},
			width:      20,
			wantHeight: 2,
			wantLines: []string{
				"[0] test label | {'t",
				"est': 'something'}",
			},
		},
	}

	for _, tc := range tt {
		count, lines := buildLines(tc.item, tc.width) // 4 -> preifx: "[x] " len=4

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
