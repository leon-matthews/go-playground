package lib

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

var companies = []Name{
	NewName("Acme Corporation"),
	NewName("Globex Corporation"),
	NewName("Soylent Corp"),
	NewName("Initech"),
	NewName("Umbrella Corporation"),
	NewName("Hooli"),
	NewName("Vehement Capital Partners"),
	NewName("Massive Dynamic"),
}

func TestName(t *testing.T) {
	name := NewName("Flintstone")

	t.Run("stringer", func(t *testing.T) {
		assert.Equal(t, "Flintstone", fmt.Sprintf("%s", name))
	})

	t.Run("length", func(t *testing.T) {
		assert.Equal(t, 10, name.Length())
	})
}

func TestReadNames(t *testing.T) {
	t.Run("missing file error", func(t *testing.T) {
		path := testDataPath("missing_file.txt")
		names, err := ReadNames(path)
		assert.Error(t, err)
		assert.Nil(t, names)
	})

	t.Run("read company names", func(t *testing.T) {
		path := testDataPath("companies.txt")
		names, err := ReadNames(path)
		require.NoError(t, err)
		assert.Equal(t, companies, names)
	})

	t.Run("ignore comments and blank lines", func(t *testing.T) {
		path := testDataPath("blanks_and_comments.txt")
		names, err := ReadNames(path)
		require.NoError(t, err)
		assert.Equal(t, companies, names)
	})
}

func TestShortestAndLongest(t *testing.T) {
	short, long := ShortestAndLongest(companies)
	assert.Equal(t, 5, short)
	assert.Equal(t, 25, long)
}
