package filter

import (
	"math/rand"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// makeHashes returns n deterministic pseudo-random 20-byte hashes.
func makeHashes(n int) [][]byte {
	r := rand.New(rand.NewSource(1))
	hashes := make([][]byte, n)
	for i := range hashes {
		h := make([]byte, 20)
		r.Read(h)
		hashes[i] = h
	}
	return hashes
}

func TestFilter(t *testing.T) {
	t.Run("New rejects non-power-of-two block counts", func(t *testing.T) {
		_, err := New(0)
		require.Error(t, err)
		_, err = New(1000)
		require.Error(t, err)

		f, err := New(1024)
		require.NoError(t, err)
		assert.Equal(t, uint64(1024), f.NumBlocks)
	})

	t.Run("BlocksForBytes rounds down to a power of two", func(t *testing.T) {
		assert.Equal(t, uint64(1), BlocksForBytes(0))
		assert.Equal(t, uint64(1), BlocksForBytes(bytesPerBlock))
		assert.Equal(t, uint64(2), BlocksForBytes(3*bytesPerBlock))
		// 16 GiB divides into exactly 2^28 blocks
		assert.Equal(t, uint64(1)<<28, BlocksForBytes(16<<30))
	})

	t.Run("Contains finds every added hash", func(t *testing.T) {
		f, err := New(4096)
		require.NoError(t, err)
		hashes := makeHashes(2000)
		for _, h := range hashes {
			f.Add(h)
		}
		for _, h := range hashes {
			assert.True(t, f.Contains(h), "added hash must be found")
		}
	})

	t.Run("never returns a false negative", func(t *testing.T) {
		f, err := New(1 << 16)
		require.NoError(t, err)
		hashes := makeHashes(50000)
		for _, h := range hashes {
			f.Add(h)
		}
		for _, h := range hashes {
			require.True(t, f.Contains(h))
		}
	})

	t.Run("false-positive rate stays low for absent hashes", func(t *testing.T) {
		f, err := New(1 << 16)
		require.NoError(t, err)
		for _, h := range makeHashes(1000) {
			f.Add(h)
		}

		// Draw fresh hashes that were not inserted
		r := rand.New(rand.NewSource(99))
		const trials = 100000
		positives := 0
		probe := make([]byte, 20)
		for range trials {
			r.Read(probe)
			if f.Contains(probe) {
				positives++
			}
		}
		assert.Less(t, positives, trials/1000, "false positives should be well under 0.1%")
	})

	t.Run("survives a write and mmap round trip", func(t *testing.T) {
		dir := t.TempDir()
		source := filepath.Join(dir, "source.db")
		require.NoError(t, os.WriteFile(source, []byte("pretend database"), 0o644))
		path := filepath.Join(dir, "test.filter")

		built, err := New(2048)
		require.NoError(t, err)
		hashes := makeHashes(500)
		for _, h := range hashes {
			built.Add(h)
		}
		built.Elements = uint64(len(hashes))
		require.NoError(t, built.Write(path, source))

		loaded, err := Open(path, source)
		require.NoError(t, err)
		defer loaded.Close()

		assert.Equal(t, uint64(500), loaded.Elements)
		assert.Equal(t, uint64(2048), loaded.NumBlocks)
		for _, h := range hashes {
			assert.True(t, loaded.Contains(h))
		}
	})

	t.Run("detects a changed source database", func(t *testing.T) {
		dir := t.TempDir()
		source := filepath.Join(dir, "source.db")
		require.NoError(t, os.WriteFile(source, []byte("original"), 0o644))
		path := filepath.Join(dir, "test.filter")

		built, err := New(1024)
		require.NoError(t, err)
		built.Add(makeHashes(1)[0])
		require.NoError(t, built.Write(path, source))

		// Rewriting the source changes its size and modification time
		require.NoError(t, os.WriteFile(source, []byte("changed content"), 0o644))

		_, err = Open(path, source)
		assert.ErrorIs(t, err, ErrStale)
	})
}
