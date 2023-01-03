package ring

type Buffer struct {
	capacity int
	head     int
	buf      []string
}

func New(capacity int) *Buffer {
	return &Buffer{
		capacity: capacity,
		head:     0,
		buf:      make([]string, capacity),
	}
}

func (rb *Buffer) Write(val string) {

	rb.buf[rb.head] = val

	rb.head = (rb.head + 1) % rb.capacity
}
