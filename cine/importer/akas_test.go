package importer

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"local.dev/cine/reader"
)

func TestImportAkas(t *testing.T) {
	ctx := context.Background()
	db := openImportDB(t)

	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)
	count, lookups, err := importAkas(ctx, tx, openFixture(t, "imdb", reader.FileTitleAkas), titleFilter{})
	require.NoError(t, err)
	assert.Equal(t, counts{read: 3, written: 3}, count)
	require.NoError(t, flushLookup(ctx, tx, "akas_region", lookups.region))
	require.NoError(t, flushLookup(ctx, tx, "akas_language", lookups.language))
	require.NoError(t, flushLookup(ctx, tx, "akas_type", lookups.akaType))
	require.NoError(t, flushLookup(ctx, tx, "akas_attribute", lookups.attribute))
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
			"SELECT name FROM akas_region WHERE id = ?", region.Int64).Scan(&regionName))
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

	t.Run("types fold into a bitmask resolvable through akas_type", func(t *testing.T) {
		var typeCount int
		require.NoError(t, db.QueryRowContext(ctx, `
			SELECT count(*) FROM akas a, akas_type t
			WHERE a.title_id = 2 AND a.ordering = 1 AND (a.types & (1 << t.id)) != 0`).Scan(&typeCount))
		assert.Equal(t, 2, typeCount) // imdbDisplay, dvd
	})

	t.Run("attributes fan out to the akas_carry_attributes junction", func(t *testing.T) {
		var attrCount int
		require.NoError(t, db.QueryRowContext(ctx,
			"SELECT count(*) FROM akas_carry_attributes WHERE title_id = 2 AND ordering = 1").Scan(&attrCount))
		assert.Equal(t, 2, attrCount) // literal title, 2014 restoration

		var attrName string
		require.NoError(t, db.QueryRowContext(ctx, `
			SELECT a.name FROM akas_carry_attributes j JOIN akas_attribute a ON a.id = j.attribute_id
			WHERE j.title_id = 2 AND j.ordering = 1 ORDER BY a.name LIMIT 1`).Scan(&attrName))
		assert.Equal(t, "2014 restoration", attrName)
	})

	t.Run("lookups populated", func(t *testing.T) {
		counts := map[string]int{}
		for _, table := range []string{"akas_region", "akas_language", "akas_type", "akas_attribute"} {
			var n int
			require.NoError(t, db.QueryRowContext(ctx, "SELECT count(*) FROM "+table).Scan(&n))
			counts[table] = n
		}
		assert.Equal(t, 2, counts["akas_region"])    // US, FR
		assert.Equal(t, 1, counts["akas_language"])  // fr
		assert.Equal(t, 2, counts["akas_type"])      // imdbDisplay, dvd
		assert.Equal(t, 2, counts["akas_attribute"]) // literal title, 2014 restoration
	})

	t.Run("a refused title takes its attributes and lookup values with it", func(t *testing.T) {
		db := openImportDB(t)
		tx, err := db.BeginTx(ctx, nil)
		require.NoError(t, err)
		count, lookups, err := importAkas(ctx, tx, openFixture(t, "imdb", reader.FileTitleAkas), allowOnly(1))
		require.NoError(t, err)
		for table, in := range map[string]*interner{
			"akas_region": lookups.region, "akas_language": lookups.language,
			"akas_type": lookups.akaType, "akas_attribute": lookups.attribute,
		} {
			require.NoError(t, flushLookup(ctx, tx, table, in))
		}
		require.NoError(t, tx.Commit())
		assert.Equal(t, counts{read: 3, written: 2}, count)

		var attributes int
		require.NoError(t, db.QueryRowContext(ctx, "SELECT count(*) FROM akas_carry_attributes").Scan(&attributes))
		assert.Zero(t, attributes, "only title 2 had any")

		got := map[string]int{}
		for _, table := range []string{"akas_region", "akas_language", "akas_type", "akas_attribute"} {
			var n int
			require.NoError(t, db.QueryRowContext(ctx, "SELECT count(*) FROM "+table).Scan(&n))
			got[table] = n
		}
		assert.Equal(t, 1, got["akas_region"]) // US only, no FR
		assert.Zero(t, got["akas_language"])   // fr was title 2's
		assert.Equal(t, 1, got["akas_type"])   // imdbDisplay only, no dvd
		assert.Zero(t, got["akas_attribute"])
	})
}
