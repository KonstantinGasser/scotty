package ring

import (
	"fmt"
	"testing"

	"golang.org/x/exp/slices"
)

func TestInsert(t *testing.T) {

	var cap uint32 = 20
	rb := New[string](cap)

	for i := 0; i < int(cap); i++ {
		rb.Write(fmt.Sprint(i))
	}

	for i := 0; i < int(cap); i++ {
		if rb.buf[i] != fmt.Sprint(i) {
			t.Fatalf("expected value: %d - got: %s", i, rb.buf[i])
		}
	}
}

func TestSeek(t *testing.T) {

	var cap uint32 = 32
	var comparable = make([]int, cap)

	rb := New[int](cap)

	for i := 0; i < int(cap); i++ {
		rb.Write(i)
		comparable[i] = i
	}

	// seek all
	all := rb.Seek(0, 32)

	if ok := slices.Compare(comparable[0:32], all); ok != 0 {
		t.Fatalf("seek-all: rb.Seek of all did not match.\nGot: %v\nWant: %v", all, comparable[0:32])
	}

	// seek last 10 inserted values
	last10Values := rb.Seek(rb.capacity-10, rb.capacity)

	if ok := slices.Compare(comparable[0:32], all); ok != 0 {
		t.Fatalf("seek-last-10: rb.Seek of the last 10 entries did not match.\nGot: %v\nWant: %v", last10Values, comparable[22:32])
	}

	// seek invalid range
	zeroSlice := rb.Seek(10, 5)
	if len(zeroSlice) != 0 {
		t.Fatalf("seek-invalid-range (n >= m): expected zero slice but got: %v", zeroSlice)
	}

	zeroSlice2 := rb.Seek(5, 33) // 33 >= rb.Cap()
	if len(zeroSlice) != 0 {
		t.Fatalf("seek-invalid-range (m >= Cap): expected zero slice but got: %v", zeroSlice2)
	}
}

func TestTail(t *testing.T) {

	var cap uint32 = 32
	var comparable = make([]int, cap)

	rb := New[int](cap)

	for i := 0; i < int(cap); i++ {
		rb.Write(i)
		comparable[i] = i
	}

	last10Values := rb.Tail(10)
	if ok := slices.Compare(last10Values, comparable[len(comparable)-10:]); ok != 0 {
		t.Fatalf("tail-last-10: Got: %v\nWant: %v", last10Values, comparable[len(comparable)-10:])
	}
}

type Log struct {
	Stream string
	Data   []byte
}

func BenchmarkWrite(b *testing.B) {

	var cap uint32 = 1 << 13
	var limit = 1000000

	var log = Log{
		Stream: "write-log-struct",
		Data: []byte(`
		{
			"level":"error",
			"ts":1673553753.136611,
			"caller":"application/structred.go:32",
			"msg":"unable to do X",
			"index":2,
			"error":"unable to do X",
			"ts":1673553753.136579,
			"stacktrace":"main.main
				/Users/konstantingasser/coffeecode/scotty/test/application/structred.go:32
				runtime.main
					/usr/local/go/src/runtime/proc.go:250"
		}
		`),
	}

	rb := New[Log](cap)

	for i := 0; i < limit; i++ {
		rb.Write(log)
	}
}
