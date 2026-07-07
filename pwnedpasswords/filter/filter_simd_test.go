//go:build goexperiment.simd && amd64

package filter

import (
	"fmt"
	"runtime"
	"simd/archsimd"
	"slices"
	"testing"

	"github.com/stretchr/testify/require"
)

// requireAVX512 skips tests and benchmarks on CPUs without AVX-512.
func requireAVX512(tb testing.TB) {
	tb.Helper()
	if !archsimd.X86.AVX512() {
		tb.Skip("CPU lacks AVX-512")
	}
}

// TestAVXMatchesScalar builds the same filter through both paths and checks
// they agree bit for bit, covering full and partial probe rounds.
func TestAVXMatchesScalar(t *testing.T) {
	requireAVX512(t)

	// makeHashes is deterministic, so the tail beyond the added prefix gives
	// queries that were never inserted.
	hashes := makeHashes(8192)
	added, absent := hashes[:4096], hashes[4096:]

	for _, probes := range []int{1, 3, 8, 10, 16, 21} {
		t.Run(fmt.Sprintf("probes=%d", probes), func(t *testing.T) {
			scalar, err := New(1024, probes)
			require.NoError(t, err)
			vector, err := New(1024, probes)
			require.NoError(t, err)

			for _, h := range added {
				scalar.Add(h)
				vector.addAVX(h)
			}
			require.True(t, slices.Equal(scalar.blocks, vector.blocks),
				"addAVX and Add must set identical bits")
			require.Equal(t, scalar.NumEntries, vector.NumEntries)

			for _, h := range added {
				require.True(t, vector.containsAVX(h), "added hash must be found")
			}
			for _, h := range absent {
				require.Equal(t, scalar.Contains(h), vector.containsAVX(h),
					"both paths must agree on misses and false positives")
			}
		})
	}
}

// locateAllScalar runs locate over every hash, folding the results into a
// checksum the compiler cannot discard.
func locateAllScalar(hashes []SHA1Hash, mask uint64, probes int) uint64 {
	var acc uint64
	for _, h := range hashes {
		base, masks := locate(h, mask, probes)
		acc ^= base ^ masks[0] ^ masks[7]
	}
	return acc
}

// locateAllAVX is locateAllScalar's twin; storing the vector keeps every lane computed.
func locateAllAVX(hashes []SHA1Hash, mask uint64, probes int) uint64 {
	var acc uint64
	var words [wordsPerBlock]uint64
	for _, h := range hashes {
		base, masks := locateAVX(h, mask, probes)
		masks.Store(&words)
		acc ^= base ^ words[0] ^ words[7]
	}
	return acc
}

// BenchmarkAVX compares the scalar and AVX-512 paths on a filter small enough
// to stay cache-resident, measuring the arithmetic rather than memory stalls.
//
// The locate pair batches the mask arithmetic inside an inlining-friendly
// helper, so it alone isolates compute from per-call overhead; divide its
// ns/op by the pool size, or read the ns/hash metric.
func BenchmarkAVX(b *testing.B) {
	requireAVX512(b)

	const blocks = 1 << 14 // 1 MiB of blocks, comfortably inside L2
	const probes = 16      // matches the suggested 8 GiB preset
	const pool = 1 << 14

	// Fill to the real filter's density of roughly 34 bits per element
	f, err := New(blocks, probes)
	require.NoError(b, err)
	hashes := makeHashes(blocks*15 + pool)
	added := hashes[:blocks*15]
	for _, h := range added {
		f.Add(h)
	}
	hits, misses := added[:pool], hashes[len(added):]

	b.Run("locate/scalar", func(b *testing.B) {
		var acc uint64
		for b.Loop() {
			acc ^= locateAllScalar(hits, f.mask, probes)
		}
		runtime.KeepAlive(acc)
		b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N)/pool, "ns/hash")
	})

	b.Run("locate/avx", func(b *testing.B) {
		var acc uint64
		for b.Loop() {
			acc ^= locateAllAVX(hits, f.mask, probes)
		}
		runtime.KeepAlive(acc)
		b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N)/pool, "ns/hash")
	})

	b.Run("contains-miss/scalar", func(b *testing.B) {
		i := 0
		var sink bool
		for b.Loop() {
			sink = f.Contains(misses[i&(pool-1)])
			i++
		}
		runtime.KeepAlive(sink)
	})

	b.Run("contains-miss/avx", func(b *testing.B) {
		i := 0
		var sink bool
		for b.Loop() {
			sink = f.containsAVX(misses[i&(pool-1)])
			i++
		}
		runtime.KeepAlive(sink)
	})

	b.Run("contains-hit/scalar", func(b *testing.B) {
		i := 0
		var sink bool
		for b.Loop() {
			sink = f.Contains(hits[i&(pool-1)])
			i++
		}
		runtime.KeepAlive(sink)
	})

	b.Run("contains-hit/avx", func(b *testing.B) {
		i := 0
		var sink bool
		for b.Loop() {
			sink = f.containsAVX(hits[i&(pool-1)])
			i++
		}
		runtime.KeepAlive(sink)
	})

	b.Run("add/scalar", func(b *testing.B) {
		g, err := New(blocks, probes)
		require.NoError(b, err)
		i := 0
		for b.Loop() {
			g.Add(hits[i&(pool-1)])
			i++
		}
	})

	b.Run("add/avx", func(b *testing.B) {
		g, err := New(blocks, probes)
		require.NoError(b, err)
		i := 0
		for b.Loop() {
			g.addAVX(hits[i&(pool-1)])
			i++
		}
	})
}
