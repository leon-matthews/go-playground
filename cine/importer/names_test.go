package importer

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const nameBasicsTSV = "nconst\tprimaryName\tbirthYear\tdeathYear\tprimaryProfession\tknownForTitles\n" +
	"nm0000001\tFred Astaire\t1899\t1987\tactor,producer\ttt0072308,tt0050419\n" +
	"nm9999999\tNo Credits\t\\N\t\\N\t\\N\t\\N\n"

func TestImportNames(t *testing.T) {
	ctx := context.Background()
	db := openImportDB(t)

	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)
	profession, err := importNames(ctx, tx, strings.NewReader(nameBasicsTSV))
	require.NoError(t, err)
	require.NoError(t, flushLookup(ctx, tx, "profession", profession))
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

	t.Run("profession lookup populated", func(t *testing.T) {
		var professions int
		require.NoError(t, db.QueryRowContext(ctx, "SELECT count(*) FROM profession").Scan(&professions))
		assert.Equal(t, 2, professions) // actor, producer
	})

	t.Run("knownForTitles left for the opt-in sub-layer", func(t *testing.T) {
		var known int
		require.NoError(t, db.QueryRowContext(ctx, "SELECT count(*) FROM name_known_for").Scan(&known))
		assert.Equal(t, 0, known)
	})
}
