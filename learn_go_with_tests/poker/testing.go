package poker

import (
	"os"
	"testing"
)

// CreateTempFile creates a real file-system file containing initialData.
// Run the returned function to remove the temporary file.
func CreateTempFile(t testing.TB, initialData string) (*os.File, func()) {
	t.Helper()
	tempfile, err := os.CreateTemp("", "db")
	if err != nil {
		t.Fatalf("could not create temp file: %v", err)
	}
	tempfile.Write([]byte(initialData))

	removeFile := func() {
		tempfile.Close()
		os.Remove(tempfile.Name())
	}

	return tempfile, removeFile
}

func NewStorageMock(league League) *StorageMock {
	return &StorageMock{
		NewInMemoryStorage(),
		make([]string, 0),
		league,
	}
}

// StorageMock provides in-memory player store
type StorageMock struct {
	*InMemoryStorage
	WinCalls []string
	league   League
}

func (s *StorageMock) GetLeague() League {
	return s.league
}

func (s *StorageMock) RecordWin(name string) {
	s.InMemoryStorage.RecordWin(name)
	s.WinCalls = append(s.WinCalls, name)
}
