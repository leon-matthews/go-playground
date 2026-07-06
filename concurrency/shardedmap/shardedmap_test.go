package shardedmap_test

import (
	"fmt"
	"sync"
	"testing"

	"local.dev/shardedmap"
)

func TestGetMissing(t *testing.T) {
	m := shardedmap.NewSharded()
	if got, ok := m.Get("nope"); ok || got != "" {
		t.Errorf(`Get("nope") = (%q, %v), want ("", false)`, got, ok)
	}
}

func TestSetThenGet(t *testing.T) {
	m := shardedmap.NewSharded()
	m.Set("color", "red")
	if got, ok := m.Get("color"); !ok || got != "red" {
		t.Errorf(`Get("color") = (%q, %v), want ("red", true)`, got, ok)
	}
}

func TestSetOverwrites(t *testing.T) {
	m := shardedmap.NewSharded()
	m.Set("color", "red")
	m.Set("color", "blue")
	if got, _ := m.Get("color"); got != "blue" {
		t.Errorf(`Get("color") = %q, want "blue"`, got)
	}
}

func TestDelete(t *testing.T) {
	m := shardedmap.NewSharded()
	m.Set("temp", "value")
	m.Delete("temp")
	if _, ok := m.Get("temp"); ok {
		t.Error(`key "temp" still present after Delete`)
	}
	// Deleting a key that was never set must be a harmless no-op.
	m.Delete("never-existed")
}

func TestLen(t *testing.T) {
	m := shardedmap.NewSharded()
	if got := m.Len(); got != 0 {
		t.Errorf("empty Len() = %d, want 0", got)
	}

	m.Set("a", "1")
	m.Set("b", "2")
	m.Set("a", "3") // overwrite, so still two distinct keys
	if got := m.Len(); got != 2 {
		t.Errorf("Len() = %d, want 2", got)
	}

	m.Delete("a")
	if got := m.Len(); got != 1 {
		t.Errorf("Len() after Delete = %d, want 1", got)
	}
}

func TestLoad(t *testing.T) {
	m := shardedmap.NewSharded()
	items := map[string]string{"one": "1", "two": "2", "three": "3"}
	m.Load(items)

	if got := m.Len(); got != len(items) {
		t.Errorf("Len() = %d, want %d", got, len(items))
	}
	for key, want := range items {
		if got, ok := m.Get(key); !ok || got != want {
			t.Errorf("Get(%q) = (%q, %v), want (%q, true)", key, got, ok, want)
		}
	}
}

// TestConcurrentAccess gives the race detector something to chew on: many
// goroutines write and read at once. Each goroutine owns a disjoint block of
// keys, so the final count is deterministic. Run with: go test -race
func TestConcurrentAccess(t *testing.T) {
	const goroutines = 50
	const perGoroutine = 1000

	m := shardedmap.NewSharded()
	var wg sync.WaitGroup
	for g := range goroutines {
		wg.Go(func() {
			for i := range perGoroutine {
				key := fmt.Sprintf("g%d-k%d", g, i)
				m.Set(key, "v")
				if _, ok := m.Get(key); !ok {
					t.Errorf("key %q missing immediately after Set", key)
				}
			}
		})
	}
	wg.Wait()

	if got, want := m.Len(), goroutines*perGoroutine; got != want {
		t.Errorf("Len() = %d, want %d", got, want)
	}
}

func ExampleSharded() {
	m := shardedmap.NewSharded()
	m.Set("greeting", "hello")
	m.Set("greeting", "hola") // overwrites the earlier value

	value, ok := m.Get("greeting")
	fmt.Println(value, ok)

	m.Delete("greeting")
	_, ok = m.Get("greeting")
	fmt.Println(ok)

	// Output:
	// hola true
	// false
}
