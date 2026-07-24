package importer

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "modernc.org/sqlite"
)

// kv is a two-column row for exercising the generic inserter.
type kv struct {
	a int64
	b string
}

func bindKV(args []any, r kv) []any { return append(args, r.a, r.b) }

func TestInsertSQL(t *testing.T) {
	assert.Equal(t,
		"INSERT INTO t (a, b) VALUES (?,?)",
		insertSQL("t", []string{"a", "b"}, 1))
	assert.Equal(t,
		"INSERT INTO t (a, b) VALUES (?,?),(?,?),(?,?)",
		insertSQL("t", []string{"a", "b"}, 3))
}

func TestBatchInserter(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "test.db")
	db, err := sql.Open("sqlite", path)
	require.NoError(t, err)
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { db.Close() })
	_, err = db.Exec("CREATE TABLE t (a INTEGER NOT NULL, b TEXT NOT NULL)")
	require.NoError(t, err)

	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)

	// Chunk size 10 with 25 rows exercises two full chunks plus a remainder
	ins, err := newSizedInserter(ctx, tx, "t", []string{"a", "b"}, bindKV, 10)
	require.NoError(t, err)

	const rows = 25
	for i := range rows {
		require.NoError(t, ins.Add(ctx, kv{a: int64(i), b: fmt.Sprintf("v%d", i)}))
	}
	require.NoError(t, ins.Flush(ctx))
	require.NoError(t, tx.Commit())

	var count int
	require.NoError(t, db.QueryRowContext(ctx, "SELECT count(*) FROM t").Scan(&count))
	assert.Equal(t, rows, count)

	var b string
	require.NoError(t, db.QueryRowContext(ctx, "SELECT b FROM t WHERE a = 17").Scan(&b))
	assert.Equal(t, "v17", b)
}

// BenchmarkInsertChunkSizes measures import throughput at a range of bind-
// parameter budgets, for an 11-column row shaped like titles, to pin
// bindParamBudget against the modernc driver's per-parameter bind cost.
func BenchmarkInsertChunkSizes(b *testing.B) {
	ctx := context.Background()
	columns := []string{"c0", "c1", "c2", "c3", "c4", "c5", "c6", "c7", "c8", "c9", "c10"}
	bind := func(args []any, r int64) []any {
		for range columns {
			args = append(args, r)
		}
		return args
	}
	const rowsPerOp = 20_000

	for _, budget := range []int{11, 22, 33, 44, 55, 88, 176} {
		chunkRows := max(1, budget/len(columns))
		b.Run(fmt.Sprintf("budget=%d/rows=%d", budget, chunkRows), func(b *testing.B) {
			db := openBench(b, columns)
			defer db.Close()

			b.ResetTimer()
			for range b.N {
				tx, err := db.BeginTx(ctx, nil)
				if err != nil {
					b.Fatal(err)
				}
				ins, err := newSizedInserter(ctx, tx, "t", columns, bind, chunkRows)
				if err != nil {
					b.Fatal(err)
				}
				for r := range rowsPerOp {
					if err := ins.Add(ctx, int64(r)); err != nil {
						b.Fatal(err)
					}
				}
				if err := ins.Flush(ctx); err != nil {
					b.Fatal(err)
				}
				if err := tx.Commit(); err != nil {
					b.Fatal(err)
				}
			}
			b.ReportMetric(float64(rowsPerOp)*float64(b.N)/b.Elapsed().Seconds(), "rows/s")
		})
	}
}

// openBench opens a fresh single-connection database with an integer table t of
// the given columns, tuned like the real bulk build.
func openBench(b *testing.B, columns []string) *sql.DB {
	b.Helper()
	path := filepath.Join(b.TempDir(), "bench.db")
	db, err := sql.Open("sqlite", path+"?_pragma=journal_mode(OFF)&_pragma=synchronous(OFF)")
	if err != nil {
		b.Fatal(err)
	}
	db.SetMaxOpenConns(1)

	defs := make([]string, len(columns))
	for i, c := range columns {
		defs[i] = c + " INTEGER"
	}
	if _, err := db.Exec("CREATE TABLE t (" + strings.Join(defs, ", ") + ")"); err != nil {
		b.Fatal(err)
	}
	return db
}
