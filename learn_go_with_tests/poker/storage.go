package poker

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
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
	database *json.Encoder
	league   League // Cache of database contents
}

func NewFileSystemStorageFromFile(path string) (*FileSystemStorage, func(), error) {
	db, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0666)

	if err != nil {
		return nil, nil, fmt.Errorf("problem opening %s %v", path, err)
	}

	closeFunc := func() {
		db.Close()
	}

	store, err := NewFileSystemStorage(db)

	if err != nil {
		return nil, nil, fmt.Errorf("problem creating file system player Store, %v ", err)
	}

	return store, closeFunc, nil
}

func NewFileSystemStorage(file *os.File) (*FileSystemStorage, error) {
	file.Seek(0, io.SeekStart)
	info, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("problem reading file: %v: %w", file.Name(), err)
	}

	// Handle empty file
	if info.Size() == 0 {
		file.Write([]byte("[]"))
		file.Seek(0, io.SeekStart)
	}

	// Read data from file
	league, err := NewLeague(file)
	if err != nil {
		return nil, fmt.Errorf("problem loading league from file: %s: %w", file.Name(), err)
	}

	storage := &FileSystemStorage{
		json.NewEncoder(&tape{file}),
		league,
	}
	return storage, nil
}

func (f *FileSystemStorage) GetLeague() League {
	sort.SliceStable(f.league, func(i, j int) bool {
		return f.league[i].Wins > f.league[j].Wins
	})
	return f.league
}

func (f *FileSystemStorage) GetPlayerScore(name string) int {
	player := f.league.Find(name)
	if player == nil {
		return 0
	}
	return player.Wins
}

func (f *FileSystemStorage) RecordWin(name string) {
	player := f.league.Find(name)

	if player == nil {
		f.league = append(f.league, Player{Name: name, Wins: 1})
	} else {
		player.Wins++
	}

	f.database.Encode(f.league)
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
