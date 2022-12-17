package store

import (
	"testing"
)

func TestQueueToRingBuffer(t *testing.T) {

	cap := uint16(4096)
	buffer := newRingBuffer(cap)

	for i := 0; i < int(cap); i++ {
		buffer.queue(i)
	}

	all := buffer.tailN(int(cap))

	for i := 0; i < int(cap); i++ {
		if i != all[i].(int) {
			t.Fatalf("[sequence check] cap=%d. got: %v, want: %d", cap, all[i], i)
		}
	}
}

func TestTailN(t *testing.T) {

	cap := uint16(4096)
	buffer := newRingBuffer(cap)

	for i := 0; i < int(cap); i++ {
		buffer.queue(i)
	}

	tail := 100
	all := buffer.tailN(100)

	for i := 0; i < tail; i++ {
		if i != all[i].(int) {
			t.Fatalf("[sequence check] cap=%d. got: %v, want: %d", cap, all[i], i)
		}
	}
}

func BenchmarkQueueToRingBuffer(b *testing.B) {

	cap := uint16(4096)
	buffer := newRingBuffer(cap)

	for i := 0; i < b.N; i++ {
		buffer.queue(i)
	}
}
