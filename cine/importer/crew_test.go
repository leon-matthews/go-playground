package importer

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"local.dev/cine/reader"
)

func TestImportCrew(t *testing.T) {
	ctx := context.Background()
	db := openImportDB(t)

	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)
	count, err := importCrew(ctx, tx, openIMDB(t, reader.FileTitleCrew))
	require.NoError(t, err)
	assert.Equal(t, counts{read: 2, written: 4}, count) // 2 input rows fan out to 4 credits
	require.NoError(t, tx.Commit())

	t.Run("directors and writers fan out to role-tagged rows", func(t *testing.T) {
		var total int
		require.NoError(t, db.QueryRowContext(ctx, "SELECT count(*) FROM titles_credit_names").Scan(&total))
		assert.Equal(t, 4, total) // (1:5 dir)(1:6 dir)(1:5 wri)(2:7 wri)
	})

	t.Run("a person who directed and wrote gets both roles", func(t *testing.T) {
		var roles int
		require.NoError(t, db.QueryRowContext(ctx,
			"SELECT count(*) FROM titles_credit_names WHERE title_id = 1 AND name_id = 5").Scan(&roles))
		assert.Equal(t, 2, roles) // role 0 and role 1
	})

	t.Run("a null director list adds no rows", func(t *testing.T) {
		var directors int
		require.NoError(t, db.QueryRowContext(ctx,
			"SELECT count(*) FROM titles_credit_names WHERE title_id = 2 AND role = 0").Scan(&directors))
		assert.Equal(t, 0, directors)

		var role int
		require.NoError(t, db.QueryRowContext(ctx,
			"SELECT role FROM titles_credit_names WHERE title_id = 2 AND name_id = 7").Scan(&role))
		assert.Equal(t, roleWriter, role)
	})
}
