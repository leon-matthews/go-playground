package main

type PlayerStorage interface {
	GetPlayerScore(name string) int
	RecordWin(name string)
	Reset(wins map[string]int)
}

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

func (s *InMemoryStorage) Reset(wins map[string]int) {
	if wins == nil {
		s.wins = map[string]int{}
	} else {
		s.wins = wins
	}
}
