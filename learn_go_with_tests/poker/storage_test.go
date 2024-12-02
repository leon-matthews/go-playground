package main

import (
	"log"
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
		store, err := NewFileSystemStorage(database)
		assert.NoError(t, err)

		got := store.GetPlayerScore("Leon")

		assert.Equal(t, 10, got)
	})

	t.Run("league from a reader", func(t *testing.T) {
		database, cleanDatabase := createTempFile(t, `[
			{"Name": "Leon", "Wins": 10},
			{"Name": "Alyson", "Wins": 33}
		]`)
		defer cleanDatabase()
		store, err := NewFileSystemStorage(database)
		assert.NoError(t, err)

		got := store.GetLeague()

		// Returned order of most wins
		want := League{
			Player{"Alyson", 33},
			Player{"Leon", 10},
		}
		assert.Equal(t, want, got)

		// Read again
		got = store.GetLeague()
		assert.Equal(t, want, got)
	})

	t.Run("sort league by number of wins", func(t *testing.T) {
		file, clean := createTempFile(t, `[
			{"Name": "Leon", "Wins": 10},
			{"Name": "Alyson", "Wins": 33}
		]`)
		defer clean()
		store, err := NewFileSystemStorage(file)
		assert.NoError(t, err)

		got := store.GetLeague()

		want := League{
			{"Alyson", 33},
			{"Leon", 10},
		}

		assert.Equal(t, want, got)
	})

	t.Run("store wins for existing player", func(t *testing.T) {
		database, cleanDatabase := createTempFile(t, `[
			{"Name": "Leon", "Wins": 10},
			{"Name": "Alyson", "Wins": 33}
		]`)
		defer cleanDatabase()
		store, err := NewFileSystemStorage(database)
		assert.NoError(t, err)

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
		store, err := NewFileSystemStorage(database)
		assert.NoError(t, err)

		store.RecordWin("Blake")

		got := store.GetPlayerScore("Blake")
		assert.Equal(t, 1, got)
	})

	t.Run("works with an empty file", func(t *testing.T) {
		file, clean := createTempFile(t, "")
		defer clean()
		store, err := NewFileSystemStorage(file)
		assert.NoError(t, err)
		log.Println(store)
	})
}

// createTempFile creates a real file-system file containing initialData.
// Run the returned function to remove the temporary file.
func createTempFile(t testing.TB, initialData string) (*os.File, func()) {
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
