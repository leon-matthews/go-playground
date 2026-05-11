package main

import (
	"io"
	"math/rand/v2"
	"os"
	"path/filepath"
	"testing"
)

const benchCacheEnv = "DUPE_BENCH_CACHE"

// loadBenchCache copies a real cache.db pointed to by DUPE_BENCH_CACHE into the
// benchmark's tempdir and opens it. Skips if the env var is unset so plain
// `go test` keeps working.
func loadBenchCache(b *testing.B) (*Cache, []string) {
	b.Helper()
	src := os.Getenv(benchCacheEnv)
	if src == "" {
		b.Skipf("set %s=/path/to/cache.db to run", benchCacheEnv)
	}
	dst := filepath.Join(b.TempDir(), "cache.db")
	if err := copyFile(src, dst); err != nil {
		b.Fatalf("copy cache: %v", err)
	}
	c, err := openCache(dst, nil)
	if err != nil {
		b.Fatalf("open cache: %v", err)
	}
	b.Cleanup(func() { _ = c.Close() })

	rows, err := c.db.Query("SELECT path FROM hashes")
	if err != nil {
		b.Fatalf("select paths: %v", err)
	}
	defer rows.Close()
	var paths []string
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err != nil {
			b.Fatalf("scan path: %v", err)
		}
		paths = append(paths, p)
	}
	if len(paths) == 0 {
		b.Fatal("cache has no rows")
	}
	b.Logf("loaded %d paths from %s", len(paths), src)
	return c, paths
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}

// BenchmarkCacheGetSequential walks paths in their stored order, approximating
// the access pattern of a fresh scan (WalkDir produces paths in directory order).
func BenchmarkCacheGetSequential(b *testing.B) {
	c, paths := loadBenchCache(b)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = c.Get(paths[i%len(paths)])
	}
}

// BenchmarkCacheGetRandom picks paths uniformly at random; worst-case for the
// SQLite page cache.
func BenchmarkCacheGetRandom(b *testing.B) {
	c, paths := loadBenchCache(b)
	rng := rand.New(rand.NewPCG(1, 2))
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = c.Get(paths[rng.IntN(len(paths))])
	}
}

// BenchmarkCacheGetParallel spreads random lookups across GOMAXPROCS goroutines,
// matching what the worker pool does on a real scan.
func BenchmarkCacheGetParallel(b *testing.B) {
	c, paths := loadBenchCache(b)
	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		rng := rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))
		for pb.Next() {
			_, _ = c.Get(paths[rng.IntN(len(paths))])
		}
	})
}
