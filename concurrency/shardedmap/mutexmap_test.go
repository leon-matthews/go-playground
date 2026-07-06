package shardedmap_test

import "sync"

// mutexMap is a plain map guarded by one mutex
// Used by benchmark for comparison purposes
type mutexMap struct {
	mu sync.Mutex
	m  map[string]string
}

func newMutexMap() *mutexMap {
	return &mutexMap{m: make(map[string]string)}
}

func (c *mutexMap) Get(key string) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	v, ok := c.m[key]
	return v, ok
}

func (c *mutexMap) Set(key, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.m[key] = value
}
