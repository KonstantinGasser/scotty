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

func TestBuildViewAfterResize(t *testing.T) {
	capacity := 62
	fill := 32
	beforePageSize := 16
	afterPagerSize := 8
	ttyWidth := 50

	store := New(uint32(capacity))

	formatter := store.NewFormatter(uint8(beforePageSize), ttyWidth)

	testLabel := "test"
	testLog := `{"hello": "world", "level": "debug", "index": {index}}`
	for i := 0; i < fill; i++ {
		log := strings.Replace(testLog, "{index}", fmt.Sprint(i), 1)
		raw := testLabel + " | " + log

		store.Insert(testLabel, len(testLabel+" | "), []byte(raw))
	}

	formatter.Load(1)

	before := formatter.String()
	wantBefore := []string{
		`>>>test | {"hello": "world", "level": "debug"...`,
		`test | {"hello": "world", "level": "debug", "...`,
		`test | {"hello": "world", "level": "debug", "...`,
		`test | {"hello": "world", "level": "debug", "...`,
		`test╭────────────────────────────────────────╮..`,
		`test│ test |                                 │..`,
		`test│ {                                      │..`,
		`test│   "hello": "world",                    │..`,
		`test│   "index": 1,                          │..`,
		`test│   "level": "debug"                     │..`,
		`test│ }                                      │..`,
		`test╰────────────────────────────────────────╯..`,
		`test | {"hello": "world", "level": "debug", "...`,
		`test | {"hello": "world", "level": "debug", "...`,
		`test | {"hello": "world", "level": "debug", "...`,
		`test | {"hello": "world", "level": "debug", "...`,
	}

	for i, line := range strings.Split(before, "\n") {
		if wantBefore[i] != line {
			t.Fatalf("[before-resize] wanted line: %q - got: %q", wantBefore[i], line)
		}
	}

	formatter.Resize(ttyWidth, uint8(afterPagerSize))

	after := formatter.String()

	wantAfter := []string{
		`>>>t╭────────────────────────────────────────╮..`,
		`test│ test |                                 │..`,
		`test│ {                                      │..`,
		`test│   "hello": "world",                    │..`,
		`test│   "index": 1,                          │..`,
		`test│   "level": "debug"                     │..`,
		`test│ }                                      │..`,
		`test╰────────────────────────────────────────╯..`,
	}

	for i, line := range strings.Split(after, "\n") {
		if wantAfter[i] != line {
			t.Fatalf("[after-resize] wanted line: %q - got: %q", wantAfter[i], line)
		}
	}
}
