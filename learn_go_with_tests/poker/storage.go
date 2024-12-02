package main

import (
	"encoding/json"
	"io"
	"sync"
)

// PlayerStorage keeps track of players and their scores
type PlayerStorage interface {
	// GetLeague builds list of all players
	GetLeague() League

	// GetPlayerScore returns number of wins by player with given name
	// Returns zero if player not found, or has no wins.
	GetPlayerScore(name string) int

	// RecordWin saves a win for the given player
	RecordWin(name string)
}

// FileSystemStorage saves data as simple JSON file
type FileSystemStorage struct {
	database io.ReadWriteSeeker
}

func (f *FileSystemStorage) GetLeague() League {
	f.database.Seek(0, io.SeekStart)
	league, _ := NewLeague(f.database)
	return league
}

func (f *FileSystemStorage) GetPlayerScore(name string) int {
	league := f.GetLeague()
	player := league.Find(name)
	if player == nil {
		return 0
	}
	return player.Wins
}

func (f *FileSystemStorage) RecordWin(name string) {
	league := f.GetLeague()
	player := league.Find(name)
	if player == nil {
		league = append(league, Player{Name: name, Wins: 1})
	} else {
		player.Wins++
	}

	f.database.Seek(0, io.SeekStart)
	json.NewEncoder(f.database).Encode(league)
}

// InMemoryStorage is as simple interface implementation for testing
type InMemoryStorage struct {
	lock sync.Mutex
	wins map[string]int
}

// NewInMemoryStorage builds a minimal storage for testing purposes
func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		wins: make(map[string]int),
	}
}

func (s *InMemoryStorage) GetLeague() League {
	league := make([]Player, 0, len(s.wins))
	for name, score := range s.wins {
		league = append(league, Player{name, score})
	}
	return league
}

func (s *InMemoryStorage) GetPlayerScore(name string) int {
	score := s.wins[name]
	return score
}

func (s *InMemoryStorage) RecordWin(name string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.wins[name]++
}

func (s *InMemoryStorage) Reset(wins map[string]int) {
	if wins == nil {
		s.wins = map[string]int{}
	} else {
		s.wins = wins
	}
}
