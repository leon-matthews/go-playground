package main

import "sync"

// PlayerStoreMemory is a cheap in-memory implementation of PlayerStore
type PlayerStoreMemory struct {
	scores map[string]int
	mu     sync.Mutex
}

func NewPlayerStoreMemory() *PlayerStoreMemory {
	return &PlayerStoreMemory{
		scores: map[string]int{},
		mu:     sync.Mutex{},
	}
}

func (s *PlayerStoreMemory) GetScore(name string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.scores[name]
}

func (s *PlayerStoreMemory) RecordWin(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.scores[name] += 1
}
