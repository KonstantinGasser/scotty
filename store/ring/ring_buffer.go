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

func (rb *Buffer[T]) Write(val T) {

	rb.buf[rb.head] = val

	rb.head = (rb.head + 1) % rb.capacity
}

func (rb *Buffer[T]) Seek(n int, m int) []T {

	if n >= m || m > int(rb.capacity) {
		return nil
	}

	return rb.buf[n:m]
}
