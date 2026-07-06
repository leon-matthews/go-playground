package shardedmap_test

import "sync"

// syncMap adapts the standard library sync.Map to our benchmark runner
// Used by benchmark for comparison purposes.
type syncMap struct {
	m sync.Map
}

func (c *syncMap) Get(key string) (string, bool) {
	v, ok := c.m.Load(key)
	if !ok {
		return "", false
	}
	return v.(string), true
}

func (c *syncMap) Set(key, value string) {
	c.m.Store(key, value)
}
