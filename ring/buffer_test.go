package ring

import (
	"bytes"
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

func makeStringN(n int, fn func(i int) string) string {

	var out = ""
	for i := 0; i < n; i++ {
		if fn != nil {
			out += fn(i)
			continue
		}
		out += fmt.Sprint(i)
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

	var buf Buffer
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
		fn    func([]byte) []byte
	}{
		{
			name:  "window last entry (N=1); buffer half full",
			n:     1,
			input: makeByteSliceN(int((1<<factor)/2), func(i int) []byte { return []byte(fmt.Sprintf("%d", i)) }),
			want:  "7", // last entry in the buffer is 7 (1<<factor/2 = 8)
			fn:    nil,
		},
		{
			name:  "window entries (N=4); buffer half full",
			n:     4,
			input: makeByteSliceN(int((1<<factor)/2), func(i int) []byte { return []byte(fmt.Sprintf("%d", i)) }),
			want:  "4567", // last 4 entries in the buffer
			fn:    nil,
		},
		{
			name:  "window entries (N=6); buffer overflow by 4",
			n:     6,
			input: makeByteSliceN(int((1<<factor)+4), func(i int) []byte { return []byte(fmt.Sprintf("%d", i)) }),
			want:  "141516171819",
			fn:    nil,
		},
		{
			name:  "window entries (N=6); buffer overflow by 4; custom func",
			n:     6,
			input: makeByteSliceN(int((1<<factor)+4), func(i int) []byte { return []byte(fmt.Sprintf("%d", i)) }),
			want:  "14,15,16,17,18,19,",
			fn: func(v []byte) []byte {
				return append(v, byte(','))
			},
		},
	}

	var buf Buffer
	var w strings.Builder

	for _, tc := range tt {

		// prepare buffer for reading
		buf = New(factor)
		for _, v := range tc.input {
			buf.Append(v)
		}

		w = strings.Builder{}

		if err := buf.Window(&w, tc.n, tc.fn); err != nil {
			t.Fatalf("[%s] unable to Window buffer, got an unexpected error: %v", tc.name, err)
		}

		if w.String() != tc.want {
			t.Fatalf("[%s] windowed string differs. want: %q, got: %q", tc.name, tc.want, w.String())
		}
	}
}

func TestScrollUp(t *testing.T) {

	tt := []struct {
		name        string
		factor      uint32
		scrollDelta int
		n           int
		input       [][]byte
		want        string
		fn          func([]byte) []byte
	}{
		{
			name:        "scroll-up (delta=3) (N=4); buffer half full",
			factor:      4, // cap => 16
			scrollDelta: 3,
			n:           2,
			// filled -> 16 / 2 = 8
			input: makeByteSliceN(int((1<<4)/2), func(i int) []byte { return []byte(fmt.Sprintf("%d", i)) }),
			want:  "34",
			fn:    nil,
		},
		{
			name:        "scroll-up (delta=1) (N=1); buffer overflowed by 2",
			factor:      2, // cap => 4
			scrollDelta: 1,
			n:           2,
			// overflowed by 16
			input: makeByteSliceN(int((1<<2)+2), func(i int) []byte { return []byte(fmt.Sprintf("%d", i)) }),
			want: makeStringN((1<<2)+2, func(i int) string {
				if i < ((1<<2)+2)-2 {
					return ""
				}
				return fmt.Sprintf("%d,", i-1)
			}), // string rep of i from 0-126 concatenated
			fn: func(v []byte) []byte {
				return append(v, byte(','))
			},
		},
		{
			name:        "scroll-up (delta=1) (N=50); buffer overflowed by 127",
			factor:      12, // cap => 4096
			scrollDelta: 1,
			n:           50,
			// overflowed by 127
			input: makeByteSliceN(int((1<<12)+127), func(i int) []byte { return []byte(fmt.Sprintf("%d", i)) }),
			want: makeStringN((1<<12)+127, func(i int) string {
				if i < ((1<<12)+127)-50 {
					return ""
				}
				return fmt.Sprintf("%d,", i-1)
			}), // string rep of i from 0-126 concatenated
			fn: func(v []byte) []byte {
				return append(v, byte(','))
			},
		},
	}

	var buf Buffer
	var w = &strings.Builder{}
	for _, tc := range tt {
		buf = New(tc.factor)

		for _, val := range tc.input {
			buf.Append(val)
		}

		if err := buf.ScrollUp(w, tc.scrollDelta, tc.n, tc.fn); err != nil {
			t.Fatalf("[%s] unable to ScrollUp buffer, got an unexpected error: %v", tc.name, err)
		}

		if w.String() != tc.want {
			t.Fatalf("[%s] scrolled-up string differs.\nwant: %q\ngot: %q", tc.name, tc.want, w.String())
		}

		w.Reset()
	}
}

func BenchmarkAppend(b *testing.B) {
	b.ReportAllocs()

	buf := New(12)

	payload := []byte(`{"level":"warn","ts":1674335370.996341,"caller":"application/structred.go:34","msg":"caution this indicates X","index":188,"ts":1674335370.996334}`)

	for i := 0; i < b.N; i++ {
		buf.Append(payload)
	}
}

func BenchmarkWindowN(b *testing.B) {

	buf := New(12)

	payload := []byte(`{"level":"warn","ts":1674335370.996341,"caller":"application/structred.go:34","msg":"caution this indicates X","index":188,"ts":1674335370.996334}`)

	for i := 0; i < (1<<12)+128; i++ {
		buf.Append(payload)
	}

	b.ReportAllocs()
	b.Run("windowing", func(bench *testing.B) {
		var w = strings.Builder{}
		size := 50 // pager height in full screen on 16'' monitor

		for i := 0; i < bench.N; i++ {
			err := buf.Window(&w, size, nil)
			if err != nil {
				bench.Fatalf("[windowing (N=50)] got an unexpected error: %v", err)
			}
		}
	})
}

func BenchmarkWindowNWithPreGrow(b *testing.B) {

	buf := New(12)

	payload := []byte(`{"level":"warn","ts":1674335370.996341,"caller":"application/structred.go:34","msg":"caution this indicates X","index":188,"ts":1674335370.996334}`)

	for i := 0; i < (1<<12)+128; i++ {
		buf.Append(payload)
	}

	b.ReportAllocs()
	b.Run("windowing-with-preGrow", func(bench *testing.B) {
		var w = strings.Builder{}
		size := 50         // pager height in full screen on 16'' monitor
		screenWidth := 200 // this will be available in the pager.Logger based on that we can determine the final max size of the string

		w.Grow((size * screenWidth) * 2)
		for i := 0; i < bench.N; i++ {
			err := buf.Window(&w, size, func(v []byte) []byte {
				if len(v) > screenWidth {
					return v[:screenWidth]
				}
				return v
			})
			if err != nil {
				bench.Fatalf("[windowing (N=50)] got an unexpected error: %v", err)
			}
		}
	})
}

func BenchmarkWindowNWithBytesBuffer(b *testing.B) {

	buf := New(12)

	payload := []byte(`{"level":"warn","ts":1674335370.996341,"caller":"application/structred.go:34","msg":"caution this indicates X","index":188,"ts":1674335370.996334}`)

	for i := 0; i < (1<<12)+128; i++ {
		buf.Append(payload)
	}

	b.ReportAllocs()
	b.Run("windowing-with-bytes_buf", func(bench *testing.B) {
		var w = bytes.Buffer{}
		size := 50 // pager height in full screen on 16'' monitor

		for i := 0; i < bench.N; i++ {
			err := buf.Window(&w, size, nil)
			if err != nil {
				bench.Fatalf("[windowing (N=50)] got an unexpected error: %v", err)
			}
		}
	})
}

func BenchmarkWindowNWithBytesBufferGrow(b *testing.B) {

	buf := New(12)

	payload := []byte(`{"level":"warn","ts":1674335370.996341,"caller":"application/structred.go:34","msg":"caution this indicates X","index":188,"ts":1674335370.996334}`)

	for i := 0; i < (1<<12)+128; i++ {
		buf.Append(payload)
	}

	b.ReportAllocs()
	b.Run("windowing-with-bytes_buf-growN", func(bench *testing.B) {
		size := 50         // pager height in full screen on 16'' monitor
		screenWidth := 200 // this will be available in the pager.Logger based on that we can determine the final max size of the string

		var w = bytes.Buffer{}
		w.Grow(size * screenWidth)

		for i := 0; i < bench.N; i++ {
			err := buf.Window(&w, size, func(v []byte) []byte {
				if len(v) > screenWidth {
					return v[:screenWidth]
				}
				return v
			})
			if err != nil {
				bench.Fatalf("[windowing (N=50)] got an unexpected error: %v", err)
			}
		}
	})
}
