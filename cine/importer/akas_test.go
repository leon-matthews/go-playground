package importer

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// akasTSV exercises a null region/language row, an original-title row with no
// types, and a row whose types and attributes are \x02-separated arrays.
const akasTSV = "titleId\tordering\ttitle\tregion\tlanguage\ttypes\tattributes\tisOriginalTitle\n" +
	"tt0000001\t1\tCarmencita\tUS\t\\N\timdbDisplay\t\\N\t0\n" +
	"tt0000001\t2\tCarmencita\t\\N\t\\N\t\\N\t\\N\t1\n" +
	"tt0000002\t1\tLe clown\tFR\tfr\timdbDisplay\x02dvd\tliteral title\x022014 restoration\t0\n"

func TestImportAkas(t *testing.T) {
	ctx := context.Background()
	db := openImportDB(t)

	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)
	count, lookups, err := importAkas(ctx, tx, strings.NewReader(akasTSV))
	require.NoError(t, err)
	assert.Equal(t, int64(3), count)
	require.NoError(t, flushLookup(ctx, tx, "regions", lookups.region))
	require.NoError(t, flushLookup(ctx, tx, "languages", lookups.language))
	require.NoError(t, flushLookup(ctx, tx, "akas_types", lookups.akaType))
	require.NoError(t, flushLookup(ctx, tx, "attributes", lookups.attribute))
	require.NoError(t, tx.Commit())

	t.Run("region is interned, an absent language is null", func(t *testing.T) {
		var (
			region   sql.NullInt64
			language sql.NullInt64
			original int64
		)
		require.NoError(t, db.QueryRowContext(ctx,
			"SELECT region, language, is_original_title FROM akas WHERE title_id = 1 AND ordering = 1").
			Scan(&region, &language, &original))
		require.True(t, region.Valid)
		assert.False(t, language.Valid) // \N language
		assert.Equal(t, int64(0), original)

		var regionName string
		require.NoError(t, db.QueryRowContext(ctx,
			"SELECT name FROM regions WHERE id = ?", region.Int64).Scan(&regionName))
		assert.Equal(t, "US", regionName)
	})

	t.Run("an original-title row sets the flag and carries no types", func(t *testing.T) {
		var (
			region   sql.NullInt64
			original int64
			types    int64
		)
		require.NoError(t, db.QueryRowContext(ctx,
			"SELECT region, is_original_title, types FROM akas WHERE title_id = 1 AND ordering = 2").
			Scan(&region, &original, &types))
		assert.False(t, region.Valid)
		assert.Equal(t, int64(1), original)
		assert.Equal(t, int64(0), types) // \N types default to 0
	})

	t.Run("types fold into a bitmask resolvable through aka_type", func(t *testing.T) {
		var typeCount int
		require.NoError(t, db.QueryRowContext(ctx, `
			SELECT count(*) FROM akas a, akas_types t
			WHERE a.title_id = 2 AND a.ordering = 1 AND (a.types & (1 << t.id)) != 0`).Scan(&typeCount))
		assert.Equal(t, 2, typeCount) // imdbDisplay, dvd
	})

	t.Run("attributes fan out to the aka_attribute junction", func(t *testing.T) {
		var attrCount int
		require.NoError(t, db.QueryRowContext(ctx,
			"SELECT count(*) FROM akas_carry_attributes WHERE title_id = 2 AND ordering = 1").Scan(&attrCount))
		assert.Equal(t, 2, attrCount) // literal title, 2014 restoration

		var attrName string
		require.NoError(t, db.QueryRowContext(ctx, `
			SELECT a.name FROM akas_carry_attributes j JOIN attributes a ON a.id = j.attribute_id
			WHERE j.title_id = 2 AND j.ordering = 1 ORDER BY a.name LIMIT 1`).Scan(&attrName))
		assert.Equal(t, "2014 restoration", attrName)
	})

	t.Run("lookups populated", func(t *testing.T) {
		counts := map[string]int{}
		for _, table := range []string{"regions", "languages", "akas_types", "attributes"} {
			var n int
			require.NoError(t, db.QueryRowContext(ctx, "SELECT count(*) FROM "+table).Scan(&n))
			counts[table] = n
		}
		assert.Equal(t, 2, counts["regions"])    // US, FR
		assert.Equal(t, 1, counts["languages"])  // fr
		assert.Equal(t, 2, counts["akas_types"]) // imdbDisplay, dvd
		assert.Equal(t, 2, counts["attributes"]) // literal title, 2014 restoration
	})
}
