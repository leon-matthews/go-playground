package main

import (
	"sync"
)

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

func (s *PlayerStoreMemory) GetScore(name string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.scores[name], nil
}

func (s *PlayerStoreMemory) RecordWin(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.scores[name] += 1
	return nil
}
