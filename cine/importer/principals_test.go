package importer

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const principalTSV = "tconst\tordering\tnconst\tcategory\tjob\tcharacters\n" +
	"tt0000001\t1\tnm0000001\tself\t\\N\t[\"Self\"]\n" +
	"tt0000001\t2\tnm0000002\tdirector\t\\N\t\\N\n" +
	"tt0000002\t1\tnm0000003\tactor\tvoice\t[\"Narrator\",\"Guard\"]\n"

func TestImportPrincipals(t *testing.T) {
	ctx := context.Background()
	db := openImportDB(t)

	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)
	lookups, err := importPrincipals(ctx, tx, strings.NewReader(principalTSV))
	require.NoError(t, err)
	require.NoError(t, flushLookup(ctx, tx, "category", lookups.category))
	require.NoError(t, flushLookup(ctx, tx, "job", lookups.job))
	require.NoError(t, tx.Commit())

	t.Run("credit keeps category and characters, null job", func(t *testing.T) {
		var (
			category int64
			job      sql.NullInt64
			chars    sql.NullString
		)
		row := db.QueryRowContext(ctx,
			"SELECT category, job, characters FROM principals WHERE title_id = 1 AND ordering = 1")
		require.NoError(t, row.Scan(&category, &job, &chars))
		assert.False(t, job.Valid)                // \N job
		assert.Equal(t, `["Self"]`, chars.String) // re-encoded JSON

		var categoryName string
		require.NoError(t, db.QueryRowContext(ctx,
			"SELECT name FROM category WHERE id = ?", category).Scan(&categoryName))
		assert.Equal(t, "self", categoryName)
	})

	t.Run("a director credit has null characters", func(t *testing.T) {
		var chars sql.NullString
		require.NoError(t, db.QueryRowContext(ctx,
			"SELECT characters FROM principals WHERE title_id = 1 AND ordering = 2").Scan(&chars))
		assert.False(t, chars.Valid)
	})

	t.Run("job interned and characters queryable via json_each", func(t *testing.T) {
		var jobName string
		require.NoError(t, db.QueryRowContext(ctx, `
			SELECT j.name FROM principals p JOIN job j ON j.id = p.job
			WHERE p.title_id = 2 AND p.ordering = 1`).Scan(&jobName))
		assert.Equal(t, "voice", jobName)

		var characters int
		require.NoError(t, db.QueryRowContext(ctx, `
			SELECT count(*) FROM principals p, json_each(p.characters)
			WHERE p.title_id = 2 AND p.ordering = 1`).Scan(&characters))
		assert.Equal(t, 2, characters) // Narrator, Guard
	})

	t.Run("lookups populated", func(t *testing.T) {
		var categories, jobs int
		require.NoError(t, db.QueryRowContext(ctx, "SELECT count(*) FROM category").Scan(&categories))
		require.NoError(t, db.QueryRowContext(ctx, "SELECT count(*) FROM job").Scan(&jobs))
		assert.Equal(t, 3, categories) // self, director, actor
		assert.Equal(t, 1, jobs)       // voice
	})
}
