package importer

import (
	"context"
	"database/sql"
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
	count, profession, err := importNames(ctx, tx, openIMDB(t, reader.FileNameBasics))
	require.NoError(t, err)
	assert.Equal(t, counts{read: 3, written: 3}, count)
	require.NoError(t, flushLookup(ctx, tx, "professions", profession))
	require.NoError(t, tx.Commit())

	t.Run("names inserted with profession bitmask and null years", func(t *testing.T) {
		var (
			name              string
			birth, death      sql.NullInt64
			primaryProfession int64
		)
		row := db.QueryRowContext(ctx,
			"SELECT primary_name, birth_year, death_year, primary_profession FROM names WHERE id = 1")
		require.NoError(t, row.Scan(&name, &birth, &death, &primaryProfession))
		assert.Equal(t, "Fred Astaire", name)
		assert.Equal(t, int64(1899), birth.Int64)
		assert.Equal(t, int64(1987), death.Int64)
		assert.Equal(t, int64(0b11), primaryProfession) // actor|producer, bits 0 and 1

		var absent int64
		require.NoError(t, db.QueryRowContext(ctx,
			"SELECT primary_profession FROM names WHERE id = 9999999").Scan(&absent))
		assert.Equal(t, int64(0), absent)
	})

	t.Run("missing primary name is stored as NULL", func(t *testing.T) {
		var name sql.NullString
		require.NoError(t, db.QueryRowContext(ctx,
			"SELECT primary_name FROM names WHERE id = 1000000").Scan(&name))
		assert.False(t, name.Valid) // not the literal "\N" the source writes
	})

	t.Run("profession lookup populated", func(t *testing.T) {
		var professions int
		require.NoError(t, db.QueryRowContext(ctx, "SELECT count(*) FROM professions").Scan(&professions))
		assert.Equal(t, 2, professions) // actor, producer
	})

	t.Run("knownForTitles has no table until the opt-in sub-layer builds it", func(t *testing.T) {
		var tables int
		require.NoError(t, db.QueryRowContext(ctx,
			"SELECT count(*) FROM sqlite_master WHERE type = 'table' AND name = 'name_known_for'").Scan(&tables))
		assert.Equal(t, 0, tables)
	})
}
