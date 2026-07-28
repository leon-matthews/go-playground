package importer

import (
	"context"
	"database/sql"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"local.dev/cine/database"
	"local.dev/cine/reader"
)

func TestImportTitles(t *testing.T) {
	ctx := context.Background()
	db := openImportDB(t)

	ratings, ratingsRead, err := loadRatings(openIMDB(t, reader.FileTitleRatings))
	require.NoError(t, err)
	require.Len(t, ratings, 2)
	assert.Equal(t, int64(2), ratingsRead)

	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)
	count, lookups, err := importTitles(ctx, tx, openIMDB(t, reader.FileTitleBasics), ratings, titleFilter{})
	require.NoError(t, err)
	assert.Equal(t, counts{read: 3, written: 3}, count)
	require.NoError(t, flushLookup(ctx, tx, "titles_type", lookups.titleType))
	require.NoError(t, flushLookup(ctx, tx, "titles_genre", lookups.genre))
	require.NoError(t, tx.Commit())

	t.Run("every row inserted", func(t *testing.T) {
		var count int
		require.NoError(t, db.QueryRowContext(ctx, "SELECT count(*) FROM titles").Scan(&count))
		assert.Equal(t, 3, count)
	})

	t.Run("ratings joined and equal original dropped to NULL", func(t *testing.T) {
		var (
			original   sql.NullString
			genres     int64
			avg, votes sql.NullInt64
		)
		row := db.QueryRowContext(ctx,
			"SELECT original_title, genres, average_rating, num_votes FROM titles WHERE id = 1")
		require.NoError(t, row.Scan(&original, &genres, &avg, &votes))
		assert.False(t, original.Valid)      // Carmencita == primary title
		assert.Equal(t, int64(0b11), genres) // Documentary|Short, bits 0 and 1
		assert.Equal(t, int64(57), avg.Int64)
		assert.Equal(t, int64(2220), votes.Int64)
	})

	t.Run("missing values become NULL and differing original kept", func(t *testing.T) {
		var (
			original sql.NullString
			start    sql.NullInt64
			genres   int64
			avg      sql.NullInt64
		)
		row := db.QueryRowContext(ctx,
			"SELECT original_title, start_year, genres, average_rating FROM titles WHERE id = 9999999")
		require.NoError(t, row.Scan(&original, &start, &genres, &avg))
		assert.Equal(t, "Original Titulo", original.String)
		assert.False(t, start.Valid)      // \N start year
		assert.Equal(t, int64(0), genres) // \N genres
		assert.False(t, avg.Valid)        // no rating row
	})

	t.Run("lookups populated and genre bit resolves by name", func(t *testing.T) {
		var titleTypes int
		require.NoError(t, db.QueryRowContext(ctx, "SELECT count(*) FROM titles_type").Scan(&titleTypes))
		assert.Equal(t, 2, titleTypes) // short, movie

		// The Matrix's bitmask must contain the Action and Sci-Fi genre bits
		var matches int
		require.NoError(t, db.QueryRowContext(ctx, `
			SELECT count(*) FROM titles_genre g, titles t
			WHERE t.id = 133093 AND t.genres & (1 << g.id) AND g.name IN ('Action', 'Sci-Fi')`).Scan(&matches))
		assert.Equal(t, 2, matches)
	})
}

func TestImportTitlesFiltered(t *testing.T) {
	ctx := context.Background()
	db := openImportDB(t)

	ratings, _, err := loadRatings(openIMDB(t, reader.FileTitleRatings))
	require.NoError(t, err)

	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)
	count, lookups, err := importTitles(ctx, tx, openIMDB(t, reader.FileTitleBasics), ratings, allowOnly(1))
	require.NoError(t, err)
	require.NoError(t, flushLookup(ctx, tx, "titles_type", lookups.titleType))
	require.NoError(t, flushLookup(ctx, tx, "titles_genre", lookups.genre))
	require.NoError(t, tx.Commit())

	t.Run("every source row is read, only the allowed ones written", func(t *testing.T) {
		assert.Equal(t, counts{read: 3, written: 1}, count)
	})

	t.Run("refused titles are absent", func(t *testing.T) {
		var ids []int64
		rows, err := db.QueryContext(ctx, "SELECT id FROM titles ORDER BY id")
		require.NoError(t, err)
		defer rows.Close()
		for rows.Next() {
			var id int64
			require.NoError(t, rows.Scan(&id))
			ids = append(ids, id)
		}
		require.NoError(t, rows.Err())
		assert.Equal(t, []int64{1}, ids)
	})

	t.Run("interners never see a refused row's values", func(t *testing.T) {
		var names []string
		rows, err := db.QueryContext(ctx, "SELECT name FROM titles_genre ORDER BY name")
		require.NoError(t, err)
		defer rows.Close()
		for rows.Next() {
			var name string
			require.NoError(t, rows.Scan(&name))
			names = append(names, name)
		}
		require.NoError(t, rows.Err())
		assert.Equal(t, []string{"Documentary", "Short"}, names, "no Action or Sci-Fi from The Matrix")

		var titleTypes int
		require.NoError(t, db.QueryRowContext(ctx, "SELECT count(*) FROM titles_type").Scan(&titleTypes))
		assert.Equal(t, 1, titleTypes, "short only, no movie")
	})
}

// openImportDB opens a fresh database with the full schema in a temp folder.
func openImportDB(t *testing.T) *sql.DB {
	t.Helper()
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "import.db")
	_, db, err := database.Open(ctx, path, true)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return db
}

// openIMDB opens a sample dataset file from testdata/imdb, keyed off the
// canonical gzip file name with its .gz suffix dropped, mirroring the reader
// package's fixtures.
func openIMDB(t *testing.T, gzName string) io.Reader {
	t.Helper()
	f, err := os.Open(filepath.Join("testdata", "imdb", strings.TrimSuffix(gzName, ".gz")))
	require.NoError(t, err)
	t.Cleanup(func() { f.Close() })
	return f
}
