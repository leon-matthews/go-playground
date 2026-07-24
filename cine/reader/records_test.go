package reader

import (
	"iter"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// collect drains an iterator into a slice, stopping at the first error.
func collect[T any](seq iter.Seq2[T, error]) ([]T, error) {
	var out []T
	for v, err := range seq {
		if err != nil {
			return out, err
		}
		out = append(out, v)
	}
	return out, nil
}

func TestReaders(t *testing.T) {
	t.Run("NameBasics", func(t *testing.T) {
		const data = "nconst\tprimaryName\tbirthYear\tdeathYear\tprimaryProfession\tknownForTitles\n" +
			"nm0000001\tFred Astaire\t1899\t1987\tactor,producer\ttt0072308,tt0050419\n" +
			"nm0000999\tStill Living\t1950\t\\N\t\\N\t\\N\n"

		got, err := collect(ReadNameBasics(strings.NewReader(data)))
		require.NoError(t, err)
		require.Len(t, got, 2)

		assert.Equal(t, "nm0000001", got[0].Nconst)
		assert.Equal(t, "Fred Astaire", got[0].PrimaryName)
		assert.Equal(t, 1899, got[0].BirthYear)
		assert.Equal(t, 1987, got[0].DeathYear)
		assert.Equal(t, []string{"actor", "producer"}, got[0].PrimaryProfession)
		assert.Equal(t, []string{"tt0072308", "tt0050419"}, got[0].KnownForTitles)

		assert.Equal(t, missing, got[1].DeathYear)
		assert.Nil(t, got[1].PrimaryProfession)
		assert.Nil(t, got[1].KnownForTitles)
	})

	t.Run("TitleAkas", func(t *testing.T) {
		const data = "titleId\tordering\ttitle\tregion\tlanguage\ttypes\tattributes\tisOriginalTitle\n" +
			"tt0000001\t1\tCarmencita\t\\N\t\\N\toriginal\t\\N\t1\n" +
			"tt0000001\t4\tCarmencita - spanyol tánc\tHU\t\\N\timdbDisplay\tliteral title\t0\n"

		got, err := collect(ReadTitleAkas(strings.NewReader(data)))
		require.NoError(t, err)
		require.Len(t, got, 2)

		assert.Equal(t, "tt0000001", got[0].TitleID)
		assert.Equal(t, 1, got[0].Ordering)
		assert.Equal(t, "Carmencita", got[0].Title)
		assert.Equal(t, "", got[0].Region)
		assert.Equal(t, []string{"original"}, got[0].Types)
		assert.Nil(t, got[0].Attributes)
		assert.True(t, got[0].IsOriginalTitle)

		assert.Equal(t, "Carmencita - spanyol tánc", got[1].Title) // UTF-8 preserved
		assert.Equal(t, "HU", got[1].Region)
		assert.Equal(t, []string{"literal title"}, got[1].Attributes)
		assert.False(t, got[1].IsOriginalTitle)
	})

	t.Run("TitleBasics", func(t *testing.T) {
		const data = "tconst\ttitleType\tprimaryTitle\toriginalTitle\tisAdult\tstartYear\tendYear\truntimeMinutes\tgenres\n" +
			"tt0000001\tshort\tCarmencita\tCarmencita\t0\t1894\t\\N\t1\tDocumentary,Short\n"

		got, err := collect(ReadTitleBasics(strings.NewReader(data)))
		require.NoError(t, err)
		require.Len(t, got, 1)

		b := got[0]
		assert.Equal(t, "tt0000001", b.Tconst)
		assert.Equal(t, "short", b.TitleType)
		assert.Equal(t, "Carmencita", b.PrimaryTitle)
		assert.Equal(t, "Carmencita", b.OriginalTitle) // kept even when equal
		assert.False(t, b.IsAdult)
		assert.Equal(t, 1894, b.StartYear)
		assert.Equal(t, missing, b.EndYear)
		assert.Equal(t, 1, b.RuntimeMinutes)
		assert.Equal(t, []string{"Documentary", "Short"}, b.Genres)
	})

	t.Run("TitleCrew", func(t *testing.T) {
		const data = "tconst\tdirectors\twriters\n" +
			"tt0000003\tnm0721526\tnm0721526\n" +
			"tt0000001\tnm0005690\t\\N\n"

		got, err := collect(ReadTitleCrew(strings.NewReader(data)))
		require.NoError(t, err)
		require.Len(t, got, 2)

		assert.Equal(t, []string{"nm0721526"}, got[0].Directors)
		assert.Equal(t, []string{"nm0721526"}, got[0].Writers)
		assert.Equal(t, []string{"nm0005690"}, got[1].Directors)
		assert.Nil(t, got[1].Writers)
	})

	t.Run("TitleEpisode", func(t *testing.T) {
		const data = "tconst\tparentTconst\tseasonNumber\tepisodeNumber\n" +
			"tt0041951\ttt0041038\t1\t9\n" +
			"tt0031458\ttt32857063\t\\N\t\\N\n"

		got, err := collect(ReadTitleEpisode(strings.NewReader(data)))
		require.NoError(t, err)
		require.Len(t, got, 2)

		assert.Equal(t, "tt0041038", got[0].ParentTconst)
		assert.Equal(t, 1, got[0].SeasonNumber)
		assert.Equal(t, 9, got[0].EpisodeNumber)

		assert.Equal(t, "tt32857063", got[1].ParentTconst) // 8-digit id kept as string
		assert.Equal(t, missing, got[1].SeasonNumber)
		assert.Equal(t, missing, got[1].EpisodeNumber)
	})

	t.Run("TitlePrincipals", func(t *testing.T) {
		const data = "tconst\tordering\tnconst\tcategory\tjob\tcharacters\n" +
			"tt0000001\t1\tnm1588970\tself\t\\N\t[\"Self\"]\n" +
			"tt0000001\t3\tnm0005690\tproducer\tproducer\t\\N\n"

		got, err := collect(ReadTitlePrincipals(strings.NewReader(data)))
		require.NoError(t, err)
		require.Len(t, got, 2)

		assert.Equal(t, 1, got[0].Ordering)
		assert.Equal(t, "self", got[0].Category)
		assert.Equal(t, "", got[0].Job)
		assert.Equal(t, []string{"Self"}, got[0].Characters)

		assert.Equal(t, "producer", got[1].Job)
		assert.Nil(t, got[1].Characters)
	})

	t.Run("TitleRatings", func(t *testing.T) {
		const data = "tconst\taverageRating\tnumVotes\n" +
			"tt0000001\t5.7\t2220\n"

		got, err := collect(ReadTitleRatings(strings.NewReader(data)))
		require.NoError(t, err)
		require.Len(t, got, 1)

		assert.Equal(t, "tt0000001", got[0].Tconst)
		assert.InDelta(t, 5.7, got[0].AverageRating, 1e-9)
		assert.Equal(t, 2220, got[0].NumVotes)
	})
}

func TestReaderErrors(t *testing.T) {
	const header = "nconst\tprimaryName\tbirthYear\tdeathYear\tprimaryProfession\tknownForTitles\n"

	t.Run("header mismatch", func(t *testing.T) {
		data := "WRONG\tprimaryName\tbirthYear\tdeathYear\tprimaryProfession\tknownForTitles\n" +
			"nm0000001\tFred Astaire\t1899\t1987\t\\N\t\\N\n"
		_, err := collect(ReadNameBasics(strings.NewReader(data)))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "header column 1")
	})

	t.Run("empty stream yields nothing", func(t *testing.T) {
		got, err := collect(ReadNameBasics(strings.NewReader("")))
		require.NoError(t, err)
		assert.Empty(t, got)
	})

	t.Run("header only yields nothing", func(t *testing.T) {
		got, err := collect(ReadNameBasics(strings.NewReader(header)))
		require.NoError(t, err)
		assert.Empty(t, got)
	})

	t.Run("bad integer reports line number", func(t *testing.T) {
		data := header + "nm0000001\tFred Astaire\t19x9\t1987\t\\N\t\\N\n"
		_, err := collect(ReadNameBasics(strings.NewReader(data)))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "line 2")
	})

	t.Run("bad row does not stop iteration", func(t *testing.T) {
		data := header +
			"nm0000001\tFirst\t1\t2\t\\N\t\\N\n" +
			"nm0000002\tShort Row\t3\n" + // wrong column count
			"nm0000003\tThird\t4\t5\t\\N\t\\N\n"

		var records []NameBasics
		var errs int
		for rec, err := range ReadNameBasics(strings.NewReader(data)) {
			if err != nil {
				errs++
				continue
			}
			records = append(records, rec)
		}
		assert.Equal(t, 1, errs)
		require.Len(t, records, 2)
		assert.Equal(t, "nm0000001", records[0].Nconst)
		assert.Equal(t, "nm0000003", records[1].Nconst)
	})

	t.Run("break stops iteration early", func(t *testing.T) {
		data := header +
			"nm0000001\tFirst\t1\t2\t\\N\t\\N\n" +
			"nm0000002\tSecond\t3\t4\t\\N\t\\N\n"

		var seen int
		for range ReadNameBasics(strings.NewReader(data)) {
			seen++
			break
		}
		assert.Equal(t, 1, seen)
	})
}
