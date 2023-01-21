package ring

import (
	"fmt"
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

	for _, tc := range tt {

		buf := New(factor)

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
