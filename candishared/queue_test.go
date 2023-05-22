package candishared

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQueue(t *testing.T) {
	q := NewQueue[string]()
	q.Push("q")
	q.Push("a")

	assert.Equal(t, 2, q.count)

	peek, _ := q.Peek()
	assert.Equal(t, "q", peek)
	pop, _ := q.Pop()
	assert.Equal(t, "q", pop)
	peek, _ = q.Peek()
	assert.Equal(t, "a", peek)
}
