package mimicry

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSummarize(t *testing.T) {
	files := []FileInfo{
		{Path: "a", Size: 10},
		{Path: "b", Size: 20},
		{Path: "c", Size: 5},
	}
	assert.Equal(t, Summary{Count: 3, Size: 35}, Summarize(files))
}

func TestSummarizeEmpty(t *testing.T) {
	assert.Equal(t, Summary{}, Summarize(nil))
}

func TestExtensionStats(t *testing.T) {
	files := []FileInfo{
		{Extension: ".txt", Size: 100},
		{Extension: ".txt", Size: 100},
		{Extension: ".txt", Size: 100},
		{Extension: ".go", Size: 50},
		{Extension: ".go", Size: 50},
		{Extension: ".go", Size: 50},
		{Extension: "", Size: 1},
	}
	// Sorted by descending count, then extension: .go and .txt tie on count (.go wins).
	want := []ExtensionStat{
		{Extension: ".go", Count: 3, Size: 150},
		{Extension: ".txt", Count: 3, Size: 300},
		{Extension: "", Count: 1, Size: 1},
	}
	assert.Equal(t, want, ExtensionStats(files))
}

func TestDuplicateGroups(t *testing.T) {
	h1 := [32]byte{1}
	h2 := [32]byte{2}
	files := []FileInfo{
		{Path: "/z/big2", Size: 1000, Hash: h1},
		{Path: "/a/big1", Size: 1000, Hash: h1},
		{Path: "/small_b", Size: 10, Hash: h2},
		{Path: "/small_a", Size: 10, Hash: h2},
		{Path: "/unique", Size: 500, Hash: [32]byte{3}},
		{Path: "/nohash", Size: 9999},
	}
	// Larger group first; files within a group ordered by path; singleton and zero-hash excluded.
	want := []DuplicateGroup{
		{Hash: h1, Size: 1000, Files: []FileInfo{
			{Path: "/a/big1", Size: 1000, Hash: h1},
			{Path: "/z/big2", Size: 1000, Hash: h1},
		}},
		{Hash: h2, Size: 10, Files: []FileInfo{
			{Path: "/small_a", Size: 10, Hash: h2},
			{Path: "/small_b", Size: 10, Hash: h2},
		}},
	}
	assert.Equal(t, want, DuplicateGroups(files))
}
