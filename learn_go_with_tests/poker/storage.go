package main

func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		wins: make(map[string]int),
	}
}

type InMemoryStorage struct {
	wins map[string]int
}

// GetPlayerScore returns number of wins by player with given name
// Returns zero if player not found, or has no wins.
func (s *InMemoryStorage) GetPlayerScore(name string) int {
	score := s.wins[name]
	return score
}

func (s *InMemoryStorage) RecordWin(name string) {
	s.wins[name]++
}
