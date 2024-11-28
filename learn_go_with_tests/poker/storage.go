package main

// PlayerStorage keeps track of players and their scores
type PlayerStorage interface {
	// GetPlayerScore returns number of wins by player with given name
	// Returns zero if player not found, or has no wins.
	GetPlayerScore(name string) int

	// RecordWin saves a win for the given player
	RecordWin(name string)

	// Reset sets the scores to the given wins map. Use nil to zero-out wins.
	Reset(wins map[string]int)
}

// NewInMemoryStorage builds a minimal storage for testing purposes
func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		wins: make(map[string]int),
	}
}

// InMemoryStorage is as simple interface implementation for testing
type InMemoryStorage struct {
	wins map[string]int
}

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
