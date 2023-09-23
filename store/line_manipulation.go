package store

import (
	"bytes"
	"strings"
	"sync"

	"github.com/KonstantinGasser/scotty/store/ring"
	"github.com/muesli/ansi"
)

const (
	indentSuffix = "| "
)

var (
	builders sync.Pool = sync.Pool{New: func() any { return bytes.NewBuffer(nil) }}
)

func lineWrap(item ring.Item, ttyWidth int) []string {

	truePrefixLen := ansi.PrintableRuneWidth(item.Raw[:item.DataPointer])
	// here we could do things better..how to avoid the string concadination?
	indent := strings.Repeat(" ", clamp(truePrefixLen-len(indentSuffix))) + indentSuffix

	if len(item.Raw[item.DataPointer:])+truePrefixLen <= ttyWidth {
		return []string{item.Raw}
	}

	// shows better results for B/op and maintains allocations (which have decreased by 1)
	// however there is no free lunch and ns/op increase on average by 100ns while dividing the B/op by 2 thou
	var builder = builders.Get().(*bytes.Buffer)
	defer func() {
		builder.Reset()
		builders.Put(builder)
	}()

	// in order to minimize runtime.growslice and
	// runtime.movemem calls we estimate how big
	// the builder's buffer has to be by using the number of characters
	// from the item paramter. However, we need to
	// take new line chars in account which is why
	// we add + len(item.Raw)/ttyWidth to the buffer size.
	// Lastly, each second+ row has an inden prefix of the
	// length of the line prefix which we need to add as well.
	builder.Grow(len(item.Raw) + len(item.Raw)/ttyWidth + (clamp(int(len(item.Raw)/ttyWidth)-1) * len(indent)))

	ansiSeqLen := len(item.Raw[:item.DataPointer]) - truePrefixLen

	var left, right = 0, ttyWidth

	// writing of the first line which includes the colores prefix
	// (colored prefix not included in second level lines)
	builder.WriteString(item.Raw[left : right+ansiSeqLen]) // special case where we can right more than the ttyWidth since ansi color sequences are not printed to the terminal as chars
	builder.WriteString("\n")

	if right+ttyWidth >= len(item.Raw)-ansiSeqLen {
		builder.WriteString(indent)
		builder.WriteString(item.Raw[right+ansiSeqLen:])

		return strings.Split(builder.String(), "\n")
	}

	left += ttyWidth + ansiSeqLen
	right += ttyWidth + ansiSeqLen

	for left < len(item.Raw) {
		builder.WriteString(indent)
		builder.WriteString(item.Raw[left : right-len(indent)])
		builder.WriteString("\n")

		left, right = left+ttyWidth-len(indent), right+ttyWidth-len(indent)
		// last bits and bytes which are left over need to be
		// written into the last line
		if right >= len(item.Raw) {
			builder.WriteString(indent)
			builder.WriteString(item.Raw[left:])
			break
		}
	}

	return strings.Split(builder.String(), "\n")
}
