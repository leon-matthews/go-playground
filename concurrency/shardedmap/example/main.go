// Command example demonstrates the shardedmap package: it seeds a map, then
// hammers it from many goroutines to show that concurrent access is safe.
package main

import (
	"fmt"
	"sync"

	"local.dev/shardedmap"
)

func main() {
	m := shardedmap.NewSharded()

	// Seed a few entries before the map is shared with any goroutine, so the
	// lock-free Load is safe to use here.
	m.Load(map[string]string{
		"language": "Go",
		"author":   "Rob Pike",
		"year":     "2009",
	})
	fmt.Printf("after Load:  %d entries\n", m.Len())

	// Fan out writers, each owning a disjoint block of keys.
	const writers = 8
	const perWriter = 10_000

	var wg sync.WaitGroup
	for w := range writers {
		wg.Go(func() {
			for i := range perWriter {
				m.Set(fmt.Sprintf("writer%d-key%d", w, i), "x")
			}
		})
	}
	wg.Wait()
	fmt.Printf("after writers: %d entries\n", m.Len())

	// Read one of the seed values back.
	if value, ok := m.Get("language"); ok {
		fmt.Printf("language = %s\n", value)
	}

	// Delete the seeds and confirm the count drops.
	for _, key := range []string{"language", "author", "year"} {
		m.Delete(key)
	}
	fmt.Printf("after deletes: %d entries\n", m.Len())
}
