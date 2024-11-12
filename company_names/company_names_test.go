package main

import (
	"github.com/stretchr/testify/assert"
	"path/filepath"
	"runtime"
	"testing"
)

func TestCountLengths(t *testing.T) {
	var dogs = []string{
		"Bear",
		"Buddy",
		"Duke",
		"Jack",
		"Lucky",
		"Moose",
		"Scout",
		"Teddy",
		"Tucker",
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
		lines, err := ReadLines(path)
		assert.Nil(t, lines)
		assert.Error(t, err)
	})

	t.Run("read company names", func(t *testing.T) {
		path := testDataPath("companies.txt")
		lines, err := ReadLines(path)
		assert.Nil(t, err)
		assert.Equal(t, companies, lines)
	})

	t.Run("ignore comments and blank lines", func(t *testing.T) {
		path := testDataPath("blanks_and_comments.txt")
		lines, err := ReadLines(path)
		assert.Nil(t, err)
		assert.Equal(t, companies, lines)
	})
}

func TestShortAndTall(t *testing.T) {
	var pangrams = []string{
		"The quick brown fox jumps over the lazy dog",
		"Sphinx of black quartz judge my vow",
		"The five boxing wizards jump quickly",
		"Jackdaws love my big sphinx of quartz",
	}

	short, long := ShortAndTall(pangrams)

	assert.Equal(t, 35, short)
	assert.Equal(t, 43, long)
}
