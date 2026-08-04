package importer

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"local.dev/cine/reader"
)

func TestImportEpisodes(t *testing.T) {
	ctx := context.Background()
	db := openImportDB(t)

	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)
	count, err := importEpisodes(ctx, tx, openIMDB(t, reader.FileTitleEpisode), titleFilter{})
	require.NoError(t, err)
	assert.Equal(t, counts{read: 3, written: 2}, count)
	require.NoError(t, tx.Commit())

	t.Run("episode keeps parent, season and number", func(t *testing.T) {
		var (
			parent          int64
			season, episode sql.NullInt64
		)
		row := db.QueryRowContext(ctx,
			"SELECT parent_id, season_number, episode_number FROM episodes WHERE id = 41038")
		require.NoError(t, row.Scan(&parent, &season, &episode))
		assert.Equal(t, int64(40021), parent)
		assert.Equal(t, int64(1), season.Int64)
		assert.Equal(t, int64(9), episode.Int64)
	})

	t.Run("missing season and number become NULL", func(t *testing.T) {
		var season, episode sql.NullInt64
		row := db.QueryRowContext(ctx,
			"SELECT season_number, episode_number FROM episodes WHERE id = 9999999")
		require.NoError(t, row.Scan(&season, &episode))
		assert.False(t, season.Valid)
		assert.False(t, episode.Valid)
	})

	t.Run("an episode with no parent is read but not written", func(t *testing.T) {
		var found int
		require.NoError(t, db.QueryRowContext(ctx,
			"SELECT count(*) FROM episodes WHERE id = 7777777").Scan(&found))
		assert.Zero(t, found)
	})

	t.Run("the filter drops an episode it did not allow", func(t *testing.T) {
		// Only the episode is checked: the filter allows one only where it kept the
		// parent, so 41038 being allowed already means 40021 was.
		db := openImportDB(t)
		tx, err := db.BeginTx(ctx, nil)
		require.NoError(t, err)
		count, err := importEpisodes(ctx, tx, openIMDB(t, reader.FileTitleEpisode), allowOnly(41038))
		require.NoError(t, err)
		require.NoError(t, tx.Commit())
		assert.Equal(t, counts{read: 3, written: 1}, count)

		var remaining int64
		require.NoError(t, db.QueryRowContext(ctx, "SELECT id FROM episodes").Scan(&remaining))
		assert.Equal(t, int64(41038), remaining)
	})
}
