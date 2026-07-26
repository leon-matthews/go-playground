package importer

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/bits-and-blooms/bitset"

	"local.dev/cine/reader"
)

// titleFilter decides which title ids a build keeps.
//
// A faithful build keeps every title and uses the zero value, which allows all
// of them, so each pass can consult a filter without first testing for one. A
// filtered build keeps only titles popular enough to have been rated, worked out
// once by buildTitleFilter before any pass runs.
type titleFilter struct {
	allowed *bitset.BitSet
}

// allows reports whether id belongs in this build.
func (f titleFilter) allows(id int64) bool {
	// A negative id wraps to an index past the end, which Test reports as unset.
	return f.allowed == nil || f.allowed.Test(uint(id))
}

// size reports how many titles the filter allows, and whether it filters at all.
func (f titleFilter) size() (uint, bool) {
	if f.allowed == nil {
		return 0, false
	}
	return f.allowed.Count(), true
}

// buildTitleFilter works out the allow-list for a filtered build from dir's
// title.episode file and the ratings already in memory.
//
// A series is kept when it carries a rating of its own or when any of its
// episodes does, and a kept series brings every one of its episodes with it. So
// a series is never stored with gaps in it, and no rating is ever discarded.
//
// Reading title.episode twice costs less than the alternative: deciding whether
// a series has a rated episode needs the whole file, and holding its 9.8 million
// id pairs in memory to avoid the second pass costs far more than the two
// bitsets do.
func buildTitleFilter(dir string, ratings map[int64]rating) (titleFilter, error) {
	path := filepath.Join(dir, reader.FileTitleEpisode)
	series, err := readKeptSeries(path, ratings)
	if err != nil {
		return titleFilter{}, err
	}
	return readAllowed(path, ratings, series)
}

// readKeptSeries opens title.episode at path for its first pass.
func readKeptSeries(path string, ratings map[int64]rating) (*bitset.BitSet, error) {
	file, err := reader.OpenGzip(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return keptSeries(file, ratings)
}

// readAllowed opens title.episode at path for its second pass.
func readAllowed(path string, ratings map[int64]rating, series *bitset.BitSet) (titleFilter, error) {
	file, err := reader.OpenGzip(path)
	if err != nil {
		return titleFilter{}, err
	}
	defer file.Close()
	return allowedTitles(file, ratings, series)
}

// keptSeries marks each series a filtered build keeps: one that is rated itself,
// or that has at least one rated episode.
func keptSeries(episodes io.Reader, ratings map[int64]rating) (*bitset.BitSet, error) {
	series := bitset.New(0)
	for record, err := range reader.ReadTitleEpisode(episodes) {
		if err != nil {
			return nil, err
		}
		parent, err := titleID(record.ParentTconst)
		if err != nil {
			return nil, err
		}
		if _, rated := ratings[int64(parent)]; rated {
			series.Set(parent)
			continue
		}
		episode, err := titleID(record.Tconst)
		if err != nil {
			return nil, err
		}
		if _, rated := ratings[int64(episode)]; rated {
			series.Set(parent)
		}
	}
	return series, nil
}

// allowedTitles builds the allow-list itself: every rated title, plus every
// episode of a kept series and the series it belongs to.
//
// An episode of a series that was not kept is cleared again, even if it was
// rated. That cannot happen under the rule above, because a rated episode always
// keeps its series, but it can once a further rule - excluding adult titles, say
// - removes a series that a rated episode belongs to. Clearing here means no
// such rule can leave an episode behind whose series is missing.
func allowedTitles(episodes io.Reader, ratings map[int64]rating, series *bitset.BitSet) (titleFilter, error) {
	allowed := bitset.New(0)
	for id := range ratings {
		// Guarded because Set allocates in proportion to its index, unlike Test.
		if id >= 0 {
			allowed.Set(uint(id))
		}
	}

	for record, err := range reader.ReadTitleEpisode(episodes) {
		if err != nil {
			return titleFilter{}, err
		}
		episode, err := titleID(record.Tconst)
		if err != nil {
			return titleFilter{}, err
		}
		parent, err := titleID(record.ParentTconst)
		if err != nil {
			return titleFilter{}, err
		}
		if series.Test(parent) {
			allowed.Set(episode)
			allowed.Set(parent)
			continue
		}
		allowed.Clear(episode)
	}
	return titleFilter{allowed: allowed}, nil
}

// titleID parses an IMDb title identifier into a bitset index.
//
// A negative id is rejected rather than passed on, because bitset.Set allocates
// memory in proportion to its index and a negative one wraps to an enormous
// index that would exhaust it.
func titleID(tconst string) (uint, error) {
	id, err := parseID(tconst)
	if err != nil {
		return 0, err
	}
	if id < 0 {
		return 0, fmt.Errorf("negative title identifier %q", tconst)
	}
	return uint(id), nil
}
