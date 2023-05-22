package candishared

import "errors"

// minQueueLen is smallest capacity that queue may have.
// Must be power of 2 for bitwise modulus: x % n == x & (n - 1).
const minQueueLen = 16

// Queue represents a single instance of the queue data structure.
type Queue[T any] struct {
	buf               []T
	head, tail, count int
}

// NewQueue constructs and returns a new Queue.
func NewQueue[T any]() *Queue[T] {
	return &Queue[T]{
		buf: make([]T, minQueueLen),
	}
}

// Len returns the number of elements currently stored in the queue.
func (q *Queue[T]) Len() int {
	return q.count
}

// Push puts an element on the end of the queue.
func (q *Queue[T]) Push(elem T) {
	if q.count == len(q.buf) {
		q.resize()
	}

	q.buf[q.tail] = elem
	q.tail = (q.tail + 1) & (len(q.buf) - 1)
	q.count++
}

// Pop Pops and returns the element from the front of the queue. If the
// queue is empty, the call will panic.
func (q *Queue[T]) Pop() (t T, err error) {
	if q.count <= 0 {
		return t, errors.New("queue: Pop() called on empty queue")
	}
	ret := q.buf[q.head]
	q.buf[q.head] = t
	// bitwise modulus
	q.head = (q.head + 1) & (len(q.buf) - 1)
	q.count--
	// Resize down if buffer 1/4 full.
	if len(q.buf) > minQueueLen && (q.count<<2) == len(q.buf) {
		q.resize()
	}
	return ret, nil
}

// Peek returns the element at the head of the queue. This call panics
// if the queue is empty.
func (q *Queue[T]) Peek() (t T, err error) {
	if q.count <= 0 {
		return t, errors.New("queue: Peek() called on empty queue")
	}
	return q.buf[q.head], nil
}

// resizes the queue to fit exactly twice its current contents
// this can result in shrinking if the queue is less than half-full
func (q *Queue[T]) resize() {
	newBuf := make([]T, q.count<<1)

	if q.tail > q.head {
		copy(newBuf, q.buf[q.head:q.tail])
	} else {
		n := copy(newBuf, q.buf[q.head:])
		copy(newBuf[n:], q.buf[:q.tail])
	}

	q.head = 0
	q.tail = q.count
	q.buf = newBuf
}
