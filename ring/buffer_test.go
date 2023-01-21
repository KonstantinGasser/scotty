package ring

import (
	"fmt"
	"strings"
	"testing"

	"golang.org/x/exp/slices"
)

func makeByteSliceN(n int, fn func(i int) []byte) [][]byte {

	var out = make([][]byte, n)
	for i := 0; i < n; i++ {
		out[i] = []byte(fn(i))
	}

	return out
}

func TestAppend(t *testing.T) {

	var factor uint32 = 2

	tt := []struct {
		name  string
		input [][]byte
		want  [][]byte
	}{
		{
			name:  "append single value",
			input: [][]byte{[]byte("hello")},
			want:  [][]byte{[]byte("hello")},
		},
		{
			name:  "append until buffer is full",
			input: makeByteSliceN((1 << factor), func(i int) []byte { return []byte(fmt.Sprintf("%d", i)) }),
			want:  makeByteSliceN((1 << factor), func(i int) []byte { return []byte(fmt.Sprintf("%d", i)) }),
		},
		{
			name:  "append until index circles",
			input: [][]byte{[]byte("0"), []byte("1"), []byte("2"), []byte("3"), []byte("4"), []byte("5")},
			want:  [][]byte{[]byte("4"), []byte("5"), []byte("2"), []byte("3")},
		},
	}

	var buf *Buffer
	for _, tc := range tt {

		buf = New(factor)

		for _, v := range tc.input {
			buf.Append(v)
		}

		for i := 0; i < len(tc.want); i++ {
			if ok := slices.Compare(buf.data[i], tc.want[i]); ok != 0 {
				t.Fatalf("[%s] buffer value differs. want: %s, got: %s", tc.name, tc.want[i], buf.data[i])
			}
		}
	}
}

func TestWindowN(t *testing.T) {

	var factor uint32 = 4

	tt := []struct {
		name  string
		n     int
		input [][]byte
		want  string
	}{
		{
			name:  "window last entry (N=1); buffer half full",
			n:     1,
			input: makeByteSliceN(int((1<<factor)/2), func(i int) []byte { return []byte(fmt.Sprintf("%d", i)) }),
			want:  "7", // last entry in the buffer is 7 (1<<factor/2 = 8)
		},
		{
			name:  "window  entries (N=4); buffer half full",
			n:     4,
			input: makeByteSliceN(int((1<<factor)/2), func(i int) []byte { return []byte(fmt.Sprintf("%d", i)) }),
			want:  "4567", // last entry in the buffer is 7 (1<<factor/2 = 8)
		},
	}

	var buf *Buffer

	var w *strings.Builder
	for _, tc := range tt {

		// prepare buffer for reading
		buf = New(factor)
		for _, v := range tc.input {
			buf.Append(v)
		}

		w = &strings.Builder{}

		if err := buf.Window(w, tc.n, nil); err != nil {
			t.Fatalf("[%s] unable to Window buffer, got an unexpected error: %v", tc.name, err)
		}

		if w.String() != tc.want {
			t.Fatalf("[%s] windowed string differs. want: %q, got: %q", tc.name, tc.want, w.String())
		}
	}
}
