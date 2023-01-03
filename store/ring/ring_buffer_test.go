package ring

import (
	"fmt"
	"testing"
)

func TestInsert(t *testing.T) {

	cap := 20
	rb := New(cap)

	for i := 0; i < cap; i++ {
		rb.Write(fmt.Sprint(i))
	}

	for i := 0; i < cap; i++ {
		if rb.buf[i] != fmt.Sprint(i) {
			t.Fatalf("expected value: %d - got: %s", i, rb.buf[i])
		}
	}
}
