package importer

import (
	"strings"
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

func TestBuildTitleFilter(t *testing.T) {
	// Two series of two episodes each: 100 is rated, 200 is not, and neither of
	// 200's episodes is rated. Film 900 is rated and has no episodes.
	const episodes = "tconst\tparentTconst\tseasonNumber\tepisodeNumber\n" +
		"tt0000101\ttt0000100\t1\t1\n" +
		"tt0000102\ttt0000100\t1\t2\n" +
		"tt0000201\ttt0000200\t1\t1\n" +
		"tt0000202\ttt0000200\t1\t2\n"

	rated := map[int64]rating{100: {average: 87, votes: 2000}, 900: {average: 57, votes: 30}}

	build := func(t *testing.T, ratings map[int64]rating) titleFilter {
		t.Helper()
		series, err := keptSeries(strings.NewReader(episodes), ratings)
		require.NoError(t, err)
		filter, err := allowedTitles(strings.NewReader(episodes), ratings, series)
		require.NoError(t, err)
		return filter
	}

	t.Run("a rated series brings all of its episodes", func(t *testing.T) {
		filter := build(t, rated)
		assert.True(t, filter.allows(100))
		assert.True(t, filter.allows(101))
		assert.True(t, filter.allows(102))
	})

	t.Run("an unrated series with no rated episodes is dropped whole", func(t *testing.T) {
		filter := build(t, rated)
		assert.False(t, filter.allows(200))
		assert.False(t, filter.allows(201))
		assert.False(t, filter.allows(202))
	})

	t.Run("a rated title without episodes is kept", func(t *testing.T) {
		filter := build(t, rated)
		assert.True(t, filter.allows(900))
	})

	t.Run("one rated episode keeps its series and its siblings", func(t *testing.T) {
		ratings := map[int64]rating{202: {average: 91, votes: 12}}
		filter := build(t, ratings)
		assert.True(t, filter.allows(200), "series adopted by its rated episode")
		assert.True(t, filter.allows(201), "unrated sibling comes along")
		assert.True(t, filter.allows(202))
	})

	t.Run("no episode survives without its series", func(t *testing.T) {
		// A series bitset that keeps nothing stands in for a later rule, such as
		// excluding adult titles, removing a series a rated episode belongs to.
		ratings := map[int64]rating{202: {average: 91, votes: 12}}
		filter, err := allowedTitles(strings.NewReader(episodes), ratings, bitset.New(0))
		require.NoError(t, err)
		assert.False(t, filter.allows(202), "rated episode cleared with its series")
		assert.False(t, filter.allows(200))
	})

	t.Run("counts the titles it allows", func(t *testing.T) {
		filter := build(t, rated)
		count, filtering := filter.size()
		assert.True(t, filtering)
		assert.Equal(t, uint(4), count, "100, 101, 102 and 900")
	})

	t.Run("a malformed identifier fails the build", func(t *testing.T) {
		const bad = "tconst\tparentTconst\tseasonNumber\tepisodeNumber\n" +
			"tt0000101\ttt-5\t1\t1\n"
		_, err := keptSeries(strings.NewReader(bad), rated)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "negative title identifier")
	})
}
