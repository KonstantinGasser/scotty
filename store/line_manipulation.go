package store

import (
	"fmt"
	"strings"

	"github.com/KonstantinGasser/scotty/store/ring"
	"github.com/muesli/ansi"
)

// buildLines takes an ring.Item as an input and based on the
// ring.Item.Raw value and the current width of the tty returns
// a slice of string containing the broken down ring.Item.Raw line
// along with the line count.
// Example:
//
//	in : label | {data: value, some: value}
//	out: 2, ["label | {data: value,", " some: value"] for width = 21
func buildLines(item ring.Item, width int, prefixOpts ...func(string) string) (int, []string) {

	prefix := fmt.Sprintf("[%d] ", item.Index())
	for _, opt := range prefixOpts {
		prefix = opt(prefix)
	}

	return breakInLines(
		item.Raw[item.DataPointer:],
		width,
		item.Raw[:item.DataPointer],
		ansi.PrintableRuneWidth(item.Raw[:item.DataPointer]),
		0,
	)

	// return breakInLines(
	// 	item.Raw[item.DataPointer:],
	// 	width,
	// 	prefix+item.Raw[:item.DataPointer],
	// 	len(prefix)+ansi.PrintableRuneWidth(item.Raw[:item.DataPointer]),
	// 	len(prefix),
	// )
}

func breakInLines(lineData string, maxWidth int, linePrefix string, printablePrefixLen int, whitespaceIndent int) (int, []string) {

	// base case:
	// prefix and line fit in tty width
	if printablePrefixLen+len(lineData) <= maxWidth {
		return 1, []string{linePrefix + lineData}
	}

	// the first line item of the returned slice
	// must start with the prefix at the beginng
	var lines []string = []string{linePrefix + lineData[:(maxWidth-printablePrefixLen)]}
	// 	=> []string{"some-prefix followed-by-the-line-of-as-much-as-we-can-write-in-the-left-over-space}

	lineData = lineData[(maxWidth - printablePrefixLen):]
	indentPrefix := strings.Repeat(" ", printablePrefixLen-2) + "| "

	if len(lineData)+printablePrefixLen <= maxWidth {
		return len(lines) + 1, append(lines, indentPrefix+lineData)
	}

	for len(lineData) >= maxWidth {
		// try to insert as much as we can (depends on the tty width)
		// of the line to the slice of line
		lineData = indentPrefix + lineData
		lines = append(lines, lineData[:maxWidth])

		// update line with what is leftover of the line
		lineData = lineData[maxWidth:]

		if len(lineData) <= maxWidth {
			return len(lines) + 1, append(lines, indentPrefix+lineData)
		}
	}

	return len(lines), lines

}

func clamp(a int) int {
	if a < 0 {
		return 0
	}
	return a
}
