package importer

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"local.dev/cine/reader"
)

func TestImportNames(t *testing.T) {
	ctx := context.Background()
	db := openImportDB(t)

	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)
	count, profession, err := importNames(ctx, tx, openFixture(t, "imdb", reader.FileNameBasics), titleFilter{})
	require.NoError(t, err)
	assert.Equal(t, counts{read: 3, written: 3}, count)
	require.NoError(t, flushLookup(ctx, tx, "names_profession", profession))
	require.NoError(t, tx.Commit())

	t.Run("names inserted with null years", func(t *testing.T) {
		var (
			name         string
			birth, death sql.NullInt64
		)
		row := db.QueryRowContext(ctx,
			"SELECT primary_name, birth_year, death_year FROM names WHERE id = 1")
		require.NoError(t, row.Scan(&name, &birth, &death))
		assert.Equal(t, "Fred Astaire", name)
		assert.Equal(t, int64(1899), birth.Int64)
		assert.Equal(t, int64(1987), death.Int64)
	})

	t.Run("professions keep IMDb's ranking, resolved by name", func(t *testing.T) {
		rows, err := db.QueryContext(ctx, `
			SELECT p.position, f.name
			FROM names_primary_professions p
			JOIN names_profession f ON f.id = p.profession_id
			WHERE p.name_id = 1
			ORDER BY p.position`)
		require.NoError(t, err)
		defer rows.Close()

		var got []string
		for rows.Next() {
			var position int64
			var profession string
			require.NoError(t, rows.Scan(&position, &profession))
			got = append(got, fmt.Sprintf("%d:%s", position, profession))
		}
		require.NoError(t, rows.Err())
		assert.Equal(t, []string{"1:actor", "2:producer"}, got)
	})

	t.Run("a \\N profession list contributes no rows", func(t *testing.T) {
		var total int
		require.NoError(t, db.QueryRowContext(ctx,
			"SELECT count(*) FROM names_primary_professions WHERE name_id = 9999999").Scan(&total))
		assert.Equal(t, 0, total)
	})

	t.Run("missing primary name is stored as NULL", func(t *testing.T) {
		var name sql.NullString
		require.NoError(t, db.QueryRowContext(ctx,
			"SELECT primary_name FROM names WHERE id = 1000000").Scan(&name))
		assert.False(t, name.Valid) // not the literal "\N" the source writes
	})

	t.Run("profession lookup populated", func(t *testing.T) {
		var professions int
		require.NoError(t, db.QueryRowContext(ctx, "SELECT count(*) FROM names_profession").Scan(&professions))
		assert.Equal(t, 2, professions) // actor, producer
	})

	t.Run("knownForTitles keeps IMDb's order, not a sorted one", func(t *testing.T) {
		rows, err := db.QueryContext(ctx,
			"SELECT position, title_id FROM names_known_for_titles WHERE name_id = 1 ORDER BY position")
		require.NoError(t, err)
		defer rows.Close()

		var got [][2]int64
		for rows.Next() {
			var position, titleID int64
			require.NoError(t, rows.Scan(&position, &titleID))
			got = append(got, [2]int64{position, titleID})
		}
		require.NoError(t, rows.Err())

		// tt0072308,tt0050419 - descending, so a sorted junction would lose this
		assert.Equal(t, [][2]int64{{1, 72308}, {2, 50419}}, got)
	})

	t.Run("a \\N known-for list contributes no rows", func(t *testing.T) {
		var total int
		require.NoError(t, db.QueryRowContext(ctx,
			"SELECT count(*) FROM names_known_for_titles WHERE name_id IN (9999999, 1000000)").Scan(&total))
		assert.Equal(t, 0, total)
	})

	t.Run("the filter drops a known-for title it did not allow", func(t *testing.T) {
		db := openImportDB(t)
		tx, err := db.BeginTx(ctx, nil)
		require.NoError(t, err)
		count, profession, err := importNames(ctx, tx, openFixture(t, "imdb", reader.FileNameBasics), allowOnly(50419))
		require.NoError(t, err)
		require.NoError(t, flushLookup(ctx, tx, "names_profession", profession))
		require.NoError(t, tx.Commit())
		assert.Equal(t, counts{read: 3, written: 3}, count, "every person is kept")

		var total int
		require.NoError(t, db.QueryRowContext(ctx,
			"SELECT count(*) FROM names_known_for_titles").Scan(&total))
		assert.Equal(t, 1, total)

		// tt0072308 was refused, so position 1 is left as a gap, not renumbered.
		var position, titleID int64
		require.NoError(t, db.QueryRowContext(ctx,
			"SELECT position, title_id FROM names_known_for_titles").Scan(&position, &titleID))
		assert.Equal(t, int64(2), position)
		assert.Equal(t, int64(50419), titleID)
	})
}
