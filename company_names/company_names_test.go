package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"path/filepath"
	"runtime"
	"testing"
)

func TestName_String(t *testing.T) {
	name := NewName("Flintstone")

	t.Run("stringer", func(t *testing.T) {
		assert.Equal(t, "Flintstone", fmt.Sprintf("%s", name))
	})

	t.Run("length", func(t *testing.T) {
		assert.Equal(t, 10, name.Length())
	})
}

func TestCountLengths(t *testing.T) {
	var dogs = []Name{
		NewName("Bear"),
		NewName("Buddy"),
		NewName("Duke"),
		NewName("Jack"),
		NewName("Lucky"),
		NewName("Moose"),
		NewName("Scout"),
		NewName("Teddy"),
		NewName("Tucker"),
	}
	counts := CountLengths(dogs)
	assert.Equal(t, counts, map[int]int{4: 3, 5: 5, 6: 1})
}

// testDataPath calculates path to a data file for testing
func testDataPath(name string) string {
	_, filename, _, _ := runtime.Caller(0)
	data_folder := filepath.Join(filepath.Dir(filename), "data")
	data_path := filepath.Join(data_folder, name)
	return data_path
}

func TestReadLines(t *testing.T) {
	var companies = []string{
		"Acme Corporation",
		"Globex Corporation",
		"Soylent Corp",
		"Initech",
		"Umbrella Corporation",
		"Hooli",
		"Vehement Capital Partners",
		"Massive Dynamic",
	}

	t.Run("missing file error", func(t *testing.T) {
		path := testDataPath("missing_file.txt")
		names, err := ReadNames(path)
		assert.Nil(t, names)
		assert.Error(t, err)
	})

	t.Run("read company names", func(t *testing.T) {
		path := testDataPath("companies.txt")
		names, err := ReadNames(path)
		assert.Nil(t, err)
		assert.Equal(t, companies, names)
	})

	t.Run("ignore comments and blank lines", func(t *testing.T) {
		path := testDataPath("blanks_and_comments.txt")
		names, err := ReadNames(path)
		assert.Nil(t, err)
		assert.Equal(t, companies, names)
	})
}
