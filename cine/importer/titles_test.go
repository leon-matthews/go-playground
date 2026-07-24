package importer

import (
	"context"
	"database/sql"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"local.dev/cine/database"
)

const ratingsTSV = "tconst\taverageRating\tnumVotes\n" +
	"tt0133093\t8.7\t2000000\n" +
	"tt0000001\t5.7\t2220\n"

const basicsTSV = "tconst\ttitleType\tprimaryTitle\toriginalTitle\tisAdult\tstartYear\tendYear\truntimeMinutes\tgenres\n" +
	"tt0000001\tshort\tCarmencita\tCarmencita\t0\t1894\t\\N\t1\tDocumentary,Short\n" +
	"tt0133093\tmovie\tThe Matrix\tThe Matrix\t0\t1999\t\\N\t136\tAction,Sci-Fi\n" +
	"tt9999999\tmovie\tTitulo\tOriginal Titulo\t0\t\\N\t\\N\t\\N\t\\N\n"

func TestImportTitles(t *testing.T) {
	ctx := context.Background()
	db := openImportDB(t)

	ratings, err := loadRatings(strings.NewReader(ratingsTSV))
	require.NoError(t, err)
	require.Len(t, ratings, 2)

	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)
	lookups, err := importTitles(ctx, tx, strings.NewReader(basicsTSV), ratings)
	require.NoError(t, err)
	require.NoError(t, flushLookup(ctx, tx, "titles_types", lookups.titleType))
	require.NoError(t, flushLookup(ctx, tx, "genres", lookups.genre))
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
		require.NoError(t, db.QueryRowContext(ctx, "SELECT count(*) FROM titles_types").Scan(&titleTypes))
		assert.Equal(t, 2, titleTypes) // short, movie

		// The Matrix's bitmask must contain the Action and Sci-Fi genre bits
		var matches int
		require.NoError(t, db.QueryRowContext(ctx, `
			SELECT count(*) FROM genres g, titles t
			WHERE t.id = 133093 AND t.genres & (1 << g.id) AND g.name IN ('Action', 'Sci-Fi')`).Scan(&matches))
		assert.Equal(t, 2, matches)
	})
}

// openImportDB opens a fresh database with the full schema in a temp folder.
func openImportDB(t *testing.T) *sql.DB {
	t.Helper()
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "import.db")
	_, db, err := database.Open(ctx, path)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return db
}
