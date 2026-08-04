package importer

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/bits-and-blooms/bitset"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTitleFilter(t *testing.T) {
	t.Run("zero value allows every title", func(t *testing.T) {
		var filter titleFilter
		assert.True(t, filter.allows(0))
		assert.True(t, filter.allows(133093))
		count, filtering := filter.size()
		assert.False(t, filtering)
		assert.Zero(t, count)
	})

	t.Run("built filter allows only its own ids", func(t *testing.T) {
		filter := titleFilter{allowed: bitset.New(0).Set(133093)}
		assert.True(t, filter.allows(133093))
		assert.False(t, filter.allows(133094))
		count, filtering := filter.size()
		assert.True(t, filtering)
		assert.Equal(t, uint(1), count)
	})

	t.Run("ids outside the bitset are refused, not fatal", func(t *testing.T) {
		filter := titleFilter{allowed: bitset.New(0).Set(1)}
		assert.False(t, filter.allows(99999999))
		assert.NotPanics(t, func() { filter.allows(-1) })
		assert.False(t, filter.allows(-1))
	})
}

// allowOnly builds a filter that allows exactly the given title ids, so a pass
// can be tested against a known allow-list without building one from a file.
func allowOnly(ids ...int64) titleFilter {
	allowed := bitset.New(0)
	for _, id := range ids {
		allowed.Set(uint(id))
	}
	return titleFilter{allowed: allowed}
}

// openFilterFixture opens one of the hand-written datasets the filter tests
// build from, which are small enough to reason about a rule against.
func openFilterFixture(t *testing.T, name string) io.Reader {
	t.Helper()
	f, err := os.Open(filepath.Join("testdata", "filter", name))
	require.NoError(t, err)
	t.Cleanup(func() { f.Close() })
	return f
}

func TestFilterBuilder(t *testing.T) {
	// In testdata/filter: series 100 is rated and clean, and episode 102 of it is
	// flagged adult. Series 200 is unrated, series 300 is rated but adult. Film 900
	// is rated and clean, 800 is rated and adult, 700 is clean and unrated.
	// Episodes 400 and 401 name no parent, 400 rated and clean, 401 unrated.
	ratings := map[int64]rating{
		100: {average: 87, votes: 2000},
		202: {average: 91, votes: 12}, // a rated episode of an unrated series
		300: {average: 62, votes: 400},
		400: {average: 73, votes: 55},
		800: {average: 41, votes: 90},
		900: {average: 57, votes: 30},
	}

	build := func(t *testing.T, options BuildOptions) titleFilter {
		t.Helper()
		builder := newFilterBuilder(options, ratings)
		require.NoError(t, builder.readBasics(openFilterFixture(t, "title.basics.tsv")))
		require.NoError(t, builder.readEpisodes(openFilterFixture(t, "title.episode.tsv")))
		return builder.filter()
	}

	t.Run("the rated rule alone ignores whether a title is adult", func(t *testing.T) {
		filter := build(t, BuildOptions{Rated: true})
		assert.True(t, filter.allows(900), "rated and clean")
		assert.True(t, filter.allows(800), "rated but adult")
		assert.False(t, filter.allows(700), "clean but unrated")
	})

	t.Run("the not-adult rule alone ignores whether a title is rated", func(t *testing.T) {
		filter := build(t, BuildOptions{NotAdult: true})
		assert.True(t, filter.allows(700), "unrated but clean")
		assert.True(t, filter.allows(900))
		assert.False(t, filter.allows(800), "adult")
	})

	t.Run("both rules together keep only titles that pass each", func(t *testing.T) {
		filter := build(t, BuildOptions{Rated: true, NotAdult: true})
		assert.True(t, filter.allows(900))
		assert.False(t, filter.allows(700), "unrated")
		assert.False(t, filter.allows(800), "adult")
	})

	t.Run("a kept series brings all of its episodes", func(t *testing.T) {
		filter := build(t, BuildOptions{Rated: true, NotAdult: true})
		assert.True(t, filter.allows(100))
		assert.True(t, filter.allows(101))
	})

	t.Run("an unrated series is dropped whole", func(t *testing.T) {
		filter := build(t, BuildOptions{Rated: true})
		assert.False(t, filter.allows(200))
		assert.False(t, filter.allows(201))
		assert.False(t, filter.allows(202))
	})

	t.Run("an adult series takes its clean episodes with it", func(t *testing.T) {
		filter := build(t, BuildOptions{NotAdult: true})
		assert.False(t, filter.allows(300))
		assert.False(t, filter.allows(301), "clean episode dropped with its adult series")
	})

	t.Run("an adult episode of a kept series survives via its parent", func(t *testing.T) {
		// Episodes are never judged by the rules, only by their parent's fate, so
		// a series is never stored with gaps in it.
		filter := build(t, BuildOptions{NotAdult: true})
		assert.True(t, filter.allows(102))
	})

	t.Run("a rated episode does not rescue its unrated series", func(t *testing.T) {
		filter := build(t, BuildOptions{Rated: true})
		assert.False(t, filter.allows(202), "rated episode cleared with its series")
	})

	t.Run("an episode with no parent keeps its own verdict", func(t *testing.T) {
		// There is no parent to inherit a fate from, so the rules that judged the
		// episode directly are the only ones that had anything to say about it.
		t.Run("kept where the rules allowed it", func(t *testing.T) {
			filter := build(t, BuildOptions{Rated: true, NotAdult: true})
			assert.True(t, filter.allows(400), "rated and clean")
		})

		t.Run("refused where they did not", func(t *testing.T) {
			filter := build(t, BuildOptions{Rated: true})
			assert.False(t, filter.allows(401), "unrated")
		})

		t.Run("and is not cleared by the absence itself", func(t *testing.T) {
			filter := build(t, BuildOptions{NotAdult: true})
			assert.True(t, filter.allows(401), "clean, and no parent refused it")
		})
	})

	t.Run("counts the titles it allows", func(t *testing.T) {
		filter := build(t, BuildOptions{Rated: true, NotAdult: true})
		count, filtering := filter.size()
		assert.True(t, filtering)
		assert.Equal(t, uint(5), count, "100, 101, 102, 400 and 900")
	})

	t.Run("a malformed identifier fails the build", func(t *testing.T) {
		builder := newFilterBuilder(BuildOptions{Rated: true}, ratings)
		err := builder.readEpisodes(openFilterFixture(t, "title.episode.malformed.tsv"))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "negative title identifier")

		t.Run("naming the row and the column it came from", func(t *testing.T) {
			assert.Contains(t, err.Error(), "row 1 (tt0000101)")
			assert.Contains(t, err.Error(), "parentTconst")
		})
	})
}

func TestBuildOptionsFiltering(t *testing.T) {
	assert.False(t, BuildOptions{}.filtering(), "the zero value filters nothing")
	assert.True(t, BuildOptions{Rated: true}.filtering())
	assert.True(t, BuildOptions{NotAdult: true}.filtering())

	// People populates tables rather than selecting rows, so it is not a filter and
	// must not make a build pay for an allow-list it has no rule to fill.
	assert.False(t, BuildOptions{People: true}.filtering())
}

func TestBuildFilter(t *testing.T) {
	t.Run("no rules means no filter and no files read", func(t *testing.T) {
		filter, err := buildFilter(t.TempDir(), BuildOptions{}, nil)
		require.NoError(t, err)
		count, filtering := filter.size()
		assert.False(t, filtering, "zero value allows every title")
		assert.Zero(t, count)
	})

	t.Run("a missing dataset file is an error", func(t *testing.T) {
		_, err := buildFilter(t.TempDir(), BuildOptions{Rated: true}, nil)
		assert.Error(t, err)
	})
}
