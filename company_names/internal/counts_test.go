package lib

import (
	"github.com/stretchr/testify/assert"
	"path/filepath"
	"runtime"
	"testing"
)

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
	assert.Equal(t, Counts{4: 3, 5: 5, 6: 1}, counts)
}

// testDataPath calculates path to a data file for testing
func testDataPath(name string) string {
	_, filename, _, _ := runtime.Caller(0)
	data_folder := filepath.Join(filepath.Dir(filename), "test_data")
	data_path := filepath.Join(data_folder, name)
	return data_path
}
