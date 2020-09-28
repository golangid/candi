package candishared

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQueue(t *testing.T) {
	q := NewQueue()
	q.Push("q")
	q.Push("a")

	assert.Equal(t, 2, q.count)
	assert.Equal(t, "q", q.Peek())
	assert.Equal(t, "q", q.Pop())
	assert.Equal(t, "a", q.Peek())
}
