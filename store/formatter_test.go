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
		`>>>t╭────────────────────────────────────────╮..`,
		`test│ test |                                 │..`,
		`test│ {                                      │..`,
		`test│   "hello": "world",                    │..`,
		`test│   "index": 1,                          │..`,
		`test│   "level": "debug"                     │..`,
		`test│ }                                      │..`,
		`test╰────────────────────────────────────────╯..`,
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
