package store

import (
	"testing"
)


func TestLineWrap(t *testing.T) {

	tt := []struct {
		name  string
		line  string
		expected string
		width int
		depth int
	}{
		{
			name:  "wrap 3 times",
			line:  "foo bar baz",
			expected: "foo\n ba\nr b\naz",
			depth: 4,
			width: 3,
		},
		{
			name: "line shorter than width",
			line: "Hello, World!",
			width: 15,
			expected: "Hello, World!",
			depth: 1,
		},
		{
			name: "line multiple times linger then width",
			line: "Lorem ipsum dolor sit amet, consectetur adipiscing elit!",
			width: 10,
			expected: "Lorem ipsu\nm dolor si\nt amet, co\nnsectetur \nadipiscing\n elit!",
			depth: 6,
		},
	}

	for _, tc := range tt {
		var depth int = 1
		lines := linewrap(&depth, tc.line, tc.width)
		if depth != tc.depth {
			t.Fatalf("[%s] expected depth of: %d; got depth: %d and lines:\n\t%q", tc.name, tc.depth, depth, lines)
		}

		if lines != tc.expected {
			t.Fatalf("[%s] expected result of: %q; got result: %q", tc.name, tc.expected, lines)
		}
	}
}
