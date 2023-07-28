package store

import (
	"fmt"
	"strings"
	"testing"
)

func TestBuildView(t *testing.T) {
	capacity := 62
	fill := 16
	pageSize := 8
	ttyWidth := 50

	store := New(uint32(capacity))

	formatter := store.NewFormatter(uint8(pageSize), ttyWidth)

	testLabel := "test"
	testLog := `{"hello": "world", "level": "debug", "index": {index}}`
	for i := 0; i < fill; i++ {
		log := strings.Replace(testLog, "{index}", fmt.Sprint(i), 1)
		raw := testLabel + " | " + log

		store.Insert(testLabel, len(testLabel+" | "), []byte(raw))
	}

	formatter.Load(1)

	formatted := formatter.String()

	want := []string{
		`[1]>╭────────────────────────────────────────╮..`,
		`[2]t│ test |                                 │..`,
		`[3]t│ {                                      │..`,
		`[4]t│   "hello": "world",                    │..`,
		`[5]t│   "index": 1,                          │..`,
		`[6]t│   "level": "debug"                     │..`,
		`[7]t│ }                                      │..`,
		`[8]t╰────────────────────────────────────────╯..`,
	}

	for i, line := range strings.Split(formatted, "\n") {
		if want[i] != line {
			t.Fatalf("wanted line: %q - got: %q", want[i], line)
		}
	}

	t.Logf(`[formatter.buildView]
DUE TO ANSI COLORS TESTING IS HARD.
PLEASE ENSURE THAT THE FOLLOWING OUTPUT
DISPLYS ALL FORMATTED JSON (FOR THE SELCTED ITEM INDEX=0).
THANKS!
<----BEGIN---->
%s
<----END---->`, formatted)
}
