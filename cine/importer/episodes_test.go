package importer

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const episodeTSV = "tconst\tparentTconst\tseasonNumber\tepisodeNumber\n" +
	"tt0041038\ttt0040021\t1\t9\n" +
	"tt9999999\ttt8888888\t\\N\t\\N\n"

func TestImportEpisodes(t *testing.T) {
	ctx := context.Background()
	db := openImportDB(t)

	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)
	count, err := importEpisodes(ctx, tx, strings.NewReader(episodeTSV))
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)
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
}
