package heap_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"heap/heap"
)

func TestPriorityQueue(t *testing.T) {
	t.Run("NewHeap queue is empty", func(t *testing.T) {
		q := heap.NewQueue[string]()
		assert.Equal(t, 0, q.Len())
	})

	t.Run("Push adds value", func(t *testing.T) {
		q := heap.NewQueue[string]()

		q.Push(1, "Hello")

		assert.Equal(t, 1, q.Len())
	})

	t.Run("Peek does not remove value", func(t *testing.T) {
		q := heap.NewQueue[string]()
		q.Push(2, "World!")
		assert.Equal(t, 1, q.Len())

		id, value, ok := q.Peek()

		assert.True(t, ok)
		assert.Equal(t, "World!", value)
		assert.Equal(t, 2, id)
		assert.Equal(t, 1, q.Len())
	})

	t.Run("peek on an empty heap returns zero vale", func(t *testing.T) {
		q := heap.NewQueue[string]()

		i, v, ok := q.Peek()

		assert.False(t, ok)
		assert.Equal(t, 0, i)
		assert.Equal(t, "", v)
	})

	t.Run("Pop removes value", func(t *testing.T) {
		q := heap.NewQueue[string]()
		q.Push(3, "Ephemeral")
		assert.Equal(t, 1, q.Len())

		id, value, ok := q.Pop()

		assert.True(t, ok)
		assert.Equal(t, "Ephemeral", value)
		assert.Equal(t, 3, id)
		assert.Equal(t, 0, q.Len())
	})

	t.Run("peek on an empty heap returns zero vale", func(t *testing.T) {
		q := heap.NewQueue[string]()

		i, v, ok := q.Pop()

		assert.False(t, ok)
		assert.Equal(t, 0, i)
		assert.Equal(t, "", v)
	})
}
