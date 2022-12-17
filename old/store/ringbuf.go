package store

type ringbuf struct {
	capacity uint16
	head     int
	buff     []interface{}
}

func newRingBuffer(cap uint16) *ringbuf {
	return &ringbuf{
		capacity: cap,
		head:     0,
		buff:     make([]interface{}, cap),
	}
}

func (rb *ringbuf) queue(p interface{}) {
	rb.buff[rb.writeIndex()] = p
	rb.head++
}

func (rb *ringbuf) tailN(n int) []interface{} {
	if len(rb.buff) < n {
		return rb.buff
	}
	return rb.buff[:n]
}

func (rb *ringbuf) writeIndex() int {
	return rb.head % int(rb.capacity)
}
