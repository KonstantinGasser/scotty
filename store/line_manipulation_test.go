package store

import (
	"testing"

	"github.com/KonstantinGasser/scotty/store/ring"
)

func TestBreakInLines(t *testing.T) {

	tt := []struct {
		name               string
		lineData           string
		maxWidth           int
		linePrefix         string
		printablePrefixLen int
		whitespaceIndent   int
		wantLines          []string
	}{
		{
			name:               "non colored prefix; wrapped multiple lines",
			lineData:           `{"level":"warn","ts":1680212791.946584,"caller":"application/structred.go:39","msg":"caution this indicates X","index":998,"ts":1680212791.946579}`,
			linePrefix:         "hello-world | ",
			maxWidth:           45,
			printablePrefixLen: len("hello-world | "),
			whitespaceIndent:   len(""),
			wantLines: []string{
				`hello-world | {"level":"warn","ts":1680212791`,
				`            | .946584,"caller":"application/s`,
				`            | tructred.go:39","msg":"caution `,
				`            | this indicates X","index":998,"`,
				`            | ts":1680212791.946579}`,
			},
		},
	}

	for _, tc := range tt {

		_, lines := breakInLines(
			tc.lineData,
			tc.maxWidth,
			tc.linePrefix,
			tc.printablePrefixLen,
			tc.whitespaceIndent,
		)

		if len(lines) != len(tc.wantLines) {
			t.Fatalf("%s: wanted %d lines, got %d lines", tc.name, len(tc.wantLines), len(lines))
		}

		for i, want := range tc.wantLines {
			if lines[i] != want {
				t.Fatalf("%s: wanted line: <%s>; got line: <%s>", tc.name, want, lines[i])
			}
		}
	}
}

/*
Current benchmark results:

goos: darwin
goarch: arm64
pkg: github.com/KonstantinGasser/scotty/store
BenchmarkBuildLines
BenchmarkBuildLines-12    	 3884172	       287.7 ns/op	     376 B/op	       8 allocs/op
*/
func BenchmarkBuildLines(b *testing.B) {

	maxWidth := 75
	item := ring.Item{
		Label:       "hello-world",
		DataPointer: 14,
		Raw:         `hello-world | {"level":"warn","ts":1680212791.946584,"caller":"application/structred.go:39","msg":"caution this indicates X","index":998,"ts":1680212791.946579}`,
	}

	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		buildLines(item, maxWidth)
	}
}
