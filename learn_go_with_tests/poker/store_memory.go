package main

// PlayerStoreMemory is a cheap in-memory implementation of PlayerStore
type PlayerStoreMemory struct {
	scores map[string]int
}

func NewPlayerStoreMemory() *PlayerStoreMemory {
	return &PlayerStoreMemory{map[string]int{}}
}

func (s *PlayerStoreMemory) GetScore(name string) int {
	return s.scores[name]
}

func (s *PlayerStoreMemory) RecordWin(name string) {
	s.scores[name] += 1
}
