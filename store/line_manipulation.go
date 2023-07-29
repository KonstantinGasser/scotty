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

	return breaklines(
		prefix+item.Raw[:item.DataPointer],
		len(prefix)+ansi.PrintableRuneWidth(item.Raw[:item.DataPointer]),
		// len(prefix)+len(item.Label)+3, // 3 => prefix format: [index]<ws>label<ws>|<ws> -> 3 for <ws>|<ws>
		item.Raw[item.DataPointer:],
		width,
		len(prefix),
	)
}

func breaklines(prefix string, escapedPrefixLen int, line string, width int, padding int) (int, []string) {

	firstLineWidth := width - escapedPrefixLen

	// base case:
	// prefix and line fit in tty width
	if escapedPrefixLen+len(line) <= width {
		return 1, []string{prefix + line}
	}

	// the first line item of the returned slice
	// must start with the prefix at the beginng
	var lines []string = []string{prefix + line[:firstLineWidth]}
	// 	=> []string{"some-prefix followed-by-the-line-of-as-much-as-we-can-write-in-the-left-over-space}

	line = line[firstLineWidth:]

	if len(line)+escapedPrefixLen <= width {
		return len(lines) + 1, append(lines, strings.Repeat(" ", padding)+prefix[padding:]+line)
	}

	for len(line) >= width {
		// try to insert as much as we can (depends on the tty width)
		// of the line to the slice of line
		line = strings.Repeat(" ", padding) + prefix[padding:] + line
		lines = append(lines, line[:width])

		// update line with what is leftover of the line
		line = line[width:]

		if len(line) <= width {
			return len(lines) + 1, append(lines, strings.Repeat(" ", padding)+prefix[padding:]+line)
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
