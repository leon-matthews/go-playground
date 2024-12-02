package main

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileSystemStorage(t *testing.T) {
	t.Run("get player score", func(t *testing.T) {
		database, cleanDatabase := createTempFile(t, `[
			{"Name": "Leon", "Wins": 10},
			{"Name": "Alyson", "Wins": 33}
		]`)
		defer cleanDatabase()
		store := NewFileSystemStorage(database)

		got := store.GetPlayerScore("Leon")

		assert.Equal(t, 10, got)
	})

	t.Run("league from a reader", func(t *testing.T) {
		database, cleanDatabase := createTempFile(t, `[
			{"Name": "Leon", "Wins": 10},
			{"Name": "Alyson", "Wins": 33}
		]`)
		defer cleanDatabase()
		store := NewFileSystemStorage(database)

		got := store.GetLeague()

		// Returned in original order
		want := League{
			Player{"Leon", 10},
			Player{"Alyson", 33},
		}
		assert.Equal(t, want, got)

		// Read again
		got = store.GetLeague()
		assert.Equal(t, want, got)
	})

	t.Run("store wins for existing player", func(t *testing.T) {
		database, cleanDatabase := createTempFile(t, `[
			{"Name": "Leon", "Wins": 10},
			{"Name": "Alyson", "Wins": 33}
		]`)
		defer cleanDatabase()
		store := NewFileSystemStorage(database)

		store.RecordWin("Leon")

		got := store.GetPlayerScore("Leon")
		assert.Equal(t, 11, got)
	})

	t.Run("store wins for new player", func(t *testing.T) {
		database, cleanDatabase := createTempFile(t, `[
			{"Name": "Leon", "Wins": 10},
			{"Name": "Alyson", "Wins": 33}
		]`)
		defer cleanDatabase()
		store := NewFileSystemStorage(database)

		store.RecordWin("Blake")

		got := store.GetPlayerScore("Blake")
		assert.Equal(t, 1, got)
	})
}

// createTempFile creates a real file-system file containing initialData.
// Run the returned function to remove the temporary file.
func createTempFile(t testing.TB, initialData string) (io.ReadWriteSeeker, func()) {
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
