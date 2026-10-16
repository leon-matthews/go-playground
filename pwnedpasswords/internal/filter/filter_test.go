package filter

import (
	"math/bits"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// makeHashes returns n deterministic pseudo-random 20-byte hashes.
func makeHashes(n int) []SHA1Hash {
	r := rand.New(rand.NewSource(1))
	hashes := make([]SHA1Hash, n)
	for i := range hashes {
		r.Read(hashes[i][:])
	}
	return hashes
}

func TestFilter(t *testing.T) {
	t.Run("BlocksForBytes finds the largest power of two that fits", func(t *testing.T) {
		assert.Equal(t, uint64(1), BlocksForBytes(0))
		assert.Equal(t, uint64(1), BlocksForBytes(bytesPerBlock))
		assert.Equal(t, uint64(2), BlocksForBytes(3*bytesPerBlock))
		// 16 GiB divides into exactly 2^28 blocks
		assert.Equal(t, uint64(1)<<28, BlocksForBytes(16<<30))
	})

	t.Run("New rejects non-power-of-two block counts", func(t *testing.T) {
		_, err := New(0, 8)
		require.Error(t, err)
		_, err = New(1000, 8)
		require.Error(t, err)

		f, err := New(1024, 8)
		require.NoError(t, err)
		assert.Equal(t, uint64(1024), f.NumBlocks)
	})

	t.Run("New rejects a non-positive probe count", func(t *testing.T) {
		_, err := New(1024, 0)
		require.Error(t, err)
	})

	t.Run("sets close to probes bits per element", func(t *testing.T) {
		// Multiply-shift probes can occasionally collide within a word, so an
		// element sets at most k bits and, averaged out, only a hair under k.
		for _, k := range []int{8, 10, 16, 21} {
			hashes := makeHashes(2000)
			total := 0
			for _, h := range hashes {
				f, err := New(8, k)
				require.NoError(t, err)
				f.Add(h)
				set := 0
				for _, w := range f.blocks {
					set += bits.OnesCount64(w)
				}
				require.LessOrEqual(t, set, k, "an element can set at most k bits")
				total += set
			}
			mean := float64(total) / float64(len(hashes))
			assert.Greater(t, mean, float64(k)-1.0, "collisions should be rare, keeping the mean near k")
		}
	})

	t.Run("Contains finds every added hash", func(t *testing.T) {
		f, err := New(4096, 16)
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
		f, err := New(1<<16, 21)
		require.NoError(t, err)
		hashes := makeHashes(50000)
		for _, h := range hashes {
			f.Add(h)
		}
		for _, h := range hashes {
			require.True(t, f.Contains(h))
		}
	})

	t.Run("false-positive rate stays low at realistic load", func(t *testing.T) {
		// Fill to eight elements per block, the load the real presets run at, so
		// probe-collision effects actually show. The old low-entropy placement sat
		// near 1 in 260 here; multiply-shift keeps the rate orders of magnitude lower.
		const blocks = 1 << 14
		const load = 8
		f, err := New(blocks, 16)
		require.NoError(t, err)
		for _, h := range makeHashes(blocks * load) {
			f.Add(h)
		}

		// Draw fresh hashes that were not inserted
		r := rand.New(rand.NewSource(99))
		const trials = 1_000_000
		positives := 0
		var probe SHA1Hash
		for range trials {
			r.Read(probe[:])
			if f.Contains(probe) {
				positives++
			}
		}
		rate := float64(positives) / trials
		assert.Less(t, rate, 1e-4, "false-positive rate must stay far below the old 1-in-260 defect")
	})

	t.Run("survives a build, write, and reopen round trip", func(t *testing.T) {
		dir := t.TempDir()
		source := filepath.Join(dir, "source.db")
		require.NoError(t, os.WriteFile(source, []byte("pretend database"), 0o644))
		path := filepath.Join(dir, "test.filter")

		built, err := New(2048, 16)
		require.NoError(t, err)
		hashes := makeHashes(500)
		for _, h := range hashes {
			built.Add(h)
		}
		require.NoError(t, built.Write(path, source))

		loaded, err := Open(path, source)
		require.NoError(t, err)
		defer loaded.Close()

		assert.Equal(t, uint64(500), loaded.NumEntries)
		assert.Equal(t, uint64(2048), loaded.NumBlocks)
		assert.Equal(t, 16, loaded.probes)
		for _, h := range hashes {
			assert.True(t, loaded.Contains(h))
		}
	})

	t.Run("detects a changed source database", func(t *testing.T) {
		dir := t.TempDir()
		source := filepath.Join(dir, "source.db")
		require.NoError(t, os.WriteFile(source, []byte("original"), 0o644))
		path := filepath.Join(dir, "test.filter")

		built, err := New(1024, 8)
		require.NoError(t, err)
		built.Add(makeHashes(1)[0])
		require.NoError(t, built.Write(path, source))

		// Rewriting the source changes its size and modification time
		require.NoError(t, os.WriteFile(source, []byte("changed content"), 0o644))

		_, err = Open(path, source)
		assert.ErrorIs(t, err, ErrStale)
	})
}

// BenchmarkContains measures Contains throughput against the real, resident
// on-disk filter, isolating the filter's per-query cost from the import
// pipeline's database work. Override the defaults with FILTER_PATH and
// FILTER_SOURCE; the benchmark skips when the files are absent or stale.
func BenchmarkContains(b *testing.B) {
	path := envOr("FILTER_PATH", "../pwnedpasswords.filter")
	source := envOr("FILTER_SOURCE", "../pwnedcache.db")

	f, err := Open(path, source)
	if err != nil {
		b.Skipf("real filter unavailable: %v", err)
	}
	defer f.Close()

	// A million distinct random hashes probe a million random blocks across the
	// 16 GiB mapping. They almost all miss, but hit and miss touch the same one
	// cache line, so this still measures the dominant cost: the random fetch.
	const pool = 1 << 20
	hashes := makeHashes(pool)

	// Warm the mapping: a freshly mmap'd region has no page-table entries even
	// when the file is fully cached, so the first touch of each page is a minor
	// fault. Fault them all in now so the timed loops measure query cost, not
	// one-off faults.
	var warm bool
	for _, h := range hashes {
		warm = f.Contains(h)
	}
	runtime.KeepAlive(warm)

	b.Run("serial", func(b *testing.B) {
		i := 0
		for b.Loop() {
			f.Contains(hashes[i&(pool-1)])
			i++
		}
	})

	b.Run("parallel", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			var sink bool
			for pb.Next() {
				sink = f.Contains(hashes[i&(pool-1)])
				i++
			}
			runtime.KeepAlive(sink)
		})
	})
}

// envOr returns the environment variable named key, or fallback when it is unset.
func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
