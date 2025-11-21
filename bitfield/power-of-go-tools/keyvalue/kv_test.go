package keyvalue_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"example/keyvalue"
)

func TestGet_KeyMissingOkayFalse(t *testing.T) {
	t.Parallel()
	s, err := keyvalue.Open("dummy.db")
	assert.NoError(t, err)
	_, ok := s.Get("test")
	assert.False(t, ok)
}

func TestGet_KeyFoundValueAndOkay(t *testing.T) {
	t.Parallel()
	s, err := keyvalue.Open("dummy.db")
	s.Set("key", "value")
	assert.NoError(t, err)
	v, ok := s.Get("key")
	assert.True(t, ok)
	assert.Equal(t, "value", v)
}

func TestOpenStore_ErrorsWhenPathUnreadable(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir() + "unreadable.store")
	_, err := os.Create(path)
	require.NoError(t, err)
	err = os.Chmod(path, 0o000)
	require.NoError(t, err)
	_, err = keyvalue.Open(path)
	assert.Error(t, err)
}

func TestSaveSavesDataPersistently(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "keyvalue.store")
	s, err := keyvalue.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	s.Set("A", "1")
	s.Set("B", "2")
	s.Set("C", "3")
	err = s.Save()
	if err != nil {
		t.Fatal(err)
	}
	s2, err := keyvalue.Open(path)
	assert.NoError(t, err)

	v, ok := s2.Get("A")
	assert.True(t, ok)
	assert.Equal(t, "1", v)

	v, ok = s2.Get("B")
	assert.True(t, ok)
	assert.Equal(t, "2", v)

	v, ok = s2.Get("C")
	assert.True(t, ok)
	assert.Equal(t, "3", v)
}

func TestSet_UpdatesValue(t *testing.T) {
	t.Parallel()
	s, err := keyvalue.Open("dummy.db")
	s.Set("key", "original")
	s.Set("key", "replacement")
	assert.NoError(t, err)
	v, ok := s.Get("key")
	assert.True(t, ok)
	assert.Equal(t, "replacement", v)
}
