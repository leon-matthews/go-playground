package shardedmap_test

import (
	"math/rand/v2"
	"strconv"
	"testing"

	"local.dev/shardedmap"
)

// Concurrency is swept via the -cpu flag (see Makefile); RunParallel runs one goroutine per CPU.
const benchKeyspace = 1_000_000

// makeKeys builds the fixed set of keys every benchmark populates and reads.
func makeKeys() []string {
	keys := make([]string, benchKeyspace)
	for i := range keys {
		keys[i] = strconv.Itoa(i)
	}
	return keys
}

func BenchmarkSharded(b *testing.B) {
	keys := makeKeys()
	m := shardedmap.NewSharded()
	for _, key := range keys {
		m.Set(key, "value")
	}

	b.Run("ro", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				m.Get(keys[rand.IntN(len(keys))])
			}
		})
	})
	b.Run("rw", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				// One draw picks both the key and the 10% write decision
				i := rand.IntN(len(keys) * 10)
				key := keys[i/10]
				if i%10 == 0 {
					m.Set(key, "value2")
				} else {
					m.Get(key)
				}
			}
		})
	})
}

func BenchmarkMutex(b *testing.B) {
	keys := makeKeys()
	m := newMutexMap()
	for _, key := range keys {
		m.Set(key, "value")
	}

	b.Run("ro", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				m.Get(keys[rand.IntN(len(keys))])
			}
		})
	})
	b.Run("rw", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				// One draw picks both the key and the 10% write decision
				i := rand.IntN(len(keys) * 10)
				key := keys[i/10]
				if i%10 == 0 {
					m.Set(key, "value2")
				} else {
					m.Get(key)
				}
			}
		})
	})
}

func BenchmarkSyncMap(b *testing.B) {
	keys := makeKeys()
	m := &syncMap{}
	for _, key := range keys {
		m.Set(key, "value")
	}

	b.Run("ro", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				m.Get(keys[rand.IntN(len(keys))])
			}
		})
	})
	b.Run("rw", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				// One draw picks both the key and the 10% write decision
				i := rand.IntN(len(keys) * 10)
				key := keys[i/10]
				if i%10 == 0 {
					m.Set(key, "value2")
				} else {
					m.Get(key)
				}
			}
		})
	})
}
