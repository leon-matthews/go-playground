package heap_test

import (
	"cmp"
	"testing"

	"priority/heap"

	"github.com/stretchr/testify/assert"
)

func TestHeap(t *testing.T) {
	t.Parallel()

	t.Run("new heap is empty", func(t *testing.T) {
		h := heap.New[int]()
		assert.Equal(t, 0, h.Len())
	})

	t.Run("push adds value", func(t *testing.T) {
		h := heap.New[int]()
		assert.Equal(t, 0, h.Len())
		h.Push(42)
		assert.Equal(t, 1, h.Len())
	})

	t.Run("peek does not remove value", func(t *testing.T) {
		h := heap.New[int]()
		h.Push(42)
		assert.Equal(t, 1, h.Len())
		assert.Equal(t, 42, h.Peek())
		assert.Equal(t, 1, h.Len())
	})

	t.Run("pop removes value", func(t *testing.T) {
		h := heap.New[int]()
		h.Push(42)
		assert.Equal(t, 1, h.Len())
		assert.Equal(t, 42, h.Pop())
		assert.Equal(t, 0, h.Len())
	})
}

type item struct {
	index int
	value int
}

func (i item) compare(b item) int {
	return cmp.Compare(i.index, b.index)
}

func TestComparableHeap(t *testing.T) {
	t.Parallel()

	t.Run("new heap is empty", func(t *testing.T) {
		h := heap.NewComparable[item](item.compare)
		assert.Equal(t, 0, h.Len())
	})

	t.Run("push adds value", func(t *testing.T) {
		h := heap.NewComparable[item](item.compare)
		assert.Equal(t, 0, h.Len())
		h.Push(item{1, 2})
		assert.Equal(t, 1, h.Len())
	})

	t.Run("peek does not remove value", func(t *testing.T) {
		h := heap.NewComparable[item](item.compare)
		h.Push(item{2, 4})
		assert.Equal(t, 1, h.Len())
		assert.Equal(t, item{2, 4}, h.Peek())
		assert.Equal(t, 1, h.Len())
	})

	t.Run("pop removes value", func(t *testing.T) {
		h := heap.NewComparable[item](item.compare)
		h.Push(item{3, 8})
		assert.Equal(t, 1, h.Len())
		assert.Equal(t, item{3, 8}, h.Pop())
		assert.Equal(t, 0, h.Len())
	})
}
