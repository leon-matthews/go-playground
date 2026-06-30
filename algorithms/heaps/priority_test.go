package heaps_test

import (
	"cmp"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"

	"local.dev/heaps"
)

func TestPriorityQueue(t *testing.T) {
	t.Run("NewHeap queue is empty", func(t *testing.T) {
		q := heaps.NewPriorityQueue[string]()
		assert.Equal(t, 0, q.Len())
	})

	t.Run("Push adds value", func(t *testing.T) {
		q := heaps.NewPriorityQueue[string]()

		q.Push(1, "Hello")

		assert.Equal(t, 1, q.Len())
	})

	t.Run("pop on an empty heap returns zero value", func(t *testing.T) {
		q := heaps.NewPriorityQueue[string]()

		i, v, ok := q.Pop()

		assert.False(t, ok)
		assert.Equal(t, 0, i)
		assert.Equal(t, "", v)
	})

	t.Run("Pop removes value", func(t *testing.T) {
		q := heaps.NewPriorityQueue[string]()
		q.Push(3, "Ephemeral")
		assert.Equal(t, 1, q.Len())

		priority, value, ok := q.Pop()

		assert.True(t, ok)
		assert.Equal(t, "Ephemeral", value)
		assert.Equal(t, 3, priority)
		assert.Equal(t, 0, q.Len())
	})

	t.Run("peek on an empty heap returns zero value", func(t *testing.T) {
		q := heaps.NewPriorityQueue[string]()

		i, v, ok := q.Peek()

		assert.False(t, ok)
		assert.Equal(t, 0, i)
		assert.Equal(t, "", v)
	})

	t.Run("Peek does not remove value", func(t *testing.T) {
		q := heaps.NewPriorityQueue[string]()
		q.Push(2, "World!")
		assert.Equal(t, 1, q.Len())

		priority, value, ok := q.Peek()

		assert.True(t, ok)
		assert.Equal(t, "World!", value)
		assert.Equal(t, 2, priority)
		assert.Equal(t, 1, q.Len())
	})

	t.Run("Values pops values by ascending priority", func(t *testing.T) {
		q := heaps.NewPriorityQueue[string]()
		q.Push(3, "black")
		q.Push(2, "of")
		q.Push(4, "quartz")
		q.Push(7, "vow")
		q.Push(5, "judge")
		q.Push(1, "Sphinx")
		q.Push(6, "my")

		s := slices.Collect(q.Values())

		// PriorityQueue is consumed
		assert.Equal(t, 0, q.Len())

		// Values have been collected in correct order
		want := []string{"Sphinx", "of", "black", "quartz", "judge", "my", "vow"}
		assert.Equal(t, want, s)
	})

	t.Run("Values partially consumes queue if interrupted", func(t *testing.T) {
		q := heaps.NewPriorityQueue[string]()
		q.Push(3, "black")
		q.Push(2, "of")
		q.Push(4, "quartz")
		q.Push(7, "vow")
		q.Push(5, "judge")
		q.Push(1, "Sphinx")
		q.Push(6, "my")

		s := make([]string, 0)
		for v := range q.Values() {
			s = append(s, v)
			if v == "quartz" {
				break
			}
		}

		// PriorityQueue is only partially consumed
		assert.Equal(t, 3, q.Len())

		// Values have been collected in correct order
		want := []string{"Sphinx", "of", "black", "quartz"}
		assert.Equal(t, want, s)
	})
}

func TestPriorityQueueStability(t *testing.T) {
	t.Run("equal priorities pop in insertion order", func(t *testing.T) {
		q := heaps.NewPriorityQueue[string]()
		q.Push(1, "first")
		q.Push(1, "second")
		q.Push(1, "third")

		want := []string{"first", "second", "third"}
		assert.Equal(t, want, slices.Collect(q.Values()))
	})

	t.Run("acts as a stable sort across duplicate priorities", func(t *testing.T) {
		const n = 1000
		q := heaps.NewPriorityQueue[int]()
		priorities := make([]int, n)
		for i := range n {
			priorities[i] = i % 5 // many ties across five priority levels
			q.Push(priorities[i], i)
		}

		// Each value is its own insertion index, so a stable sort of those indices by
		// priority is exactly the order a FIFO-stable queue must reproduce.
		want := make([]int, n)
		for i := range want {
			want[i] = i
		}
		slices.SortStableFunc(want, func(a, b int) int {
			return cmp.Compare(priorities[a], priorities[b])
		})

		assert.Equal(t, want, slices.Collect(q.Values()))
	})
}
