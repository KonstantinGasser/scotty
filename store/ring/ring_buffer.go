package ring

type Buffer[T any] struct {
	capacity uint32
	head     uint32
	buf      []T
}

func New[T any](capacity uint32) *Buffer[T] {
	return &Buffer[T]{
		capacity: capacity,
		head:     0,
		buf:      make([]T, capacity),
	}
}

func (rb *Buffer[T]) Cap() uint32 {
	return rb.capacity
}

func (rb *Buffer[T]) Write(val T) {

	rb.buf[rb.head] = val

	rb.head = (rb.head + 1) % rb.capacity
}

func (rb *Buffer[T]) Seek(n uint32, m uint32) []T {

	if n >= m || m > rb.Cap() {
		return []T{}
	}

	return rb.buf[n:m]
}

func (rb *Buffer[T]) Tail(n uint32) []T {

	if n >= rb.Cap() || n <= 0 {
		return []T{}
	}

	return rb.buf[rb.Cap()-n:]
}
