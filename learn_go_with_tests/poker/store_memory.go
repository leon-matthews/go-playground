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

func (s *PlayerStoreMemory) League() (League, error) {
	var league League
	for name, wins := range s.scores {
		league = append(league, Player{name, wins})
	}
	return league, nil
}

func (s *PlayerStoreMemory) Score(name string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.scores[name], nil
}

func (s *PlayerStoreMemory) SetScore(name string, score int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.scores[name] = score
	return nil
}

func (s *PlayerStoreMemory) RecordWin(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.scores[name]++
	return nil
}
