package store

import (
	"testing"
)

func TestLineWrap(t *testing.T) {

	tt := []struct {
		name  string
		line  string
		width int
		depth int
	}{
		{
			name:  "wrap 3 times",
			line:  "foo bar baz",
			depth: 3,
			width: 3,
		},
	}

	for _, tc := range tt {
		var depth int
		lines := linewrap(&depth, tc.line, tc.width)
		if depth != tc.depth {
			t.Fatalf("[%s] expected depth of: %d; got depth: %d and lines:\n\t%q", tc.name, tc.depth, depth, lines)
		}
	}
}
