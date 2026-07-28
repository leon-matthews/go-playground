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
// A zero value allows every title. A built filter holds the titles every enabled
// rule accepted, plus the episodes of each series it kept.
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

// buildFilter works out the allow-list for a filtered build, reading
// title.basics for the rules and then title.episode for the parent cascade.
func buildFilter(dir string, options BuildOptions, ratings map[int64]rating) (titleFilter, error) {
	if !options.filtering() {
		return titleFilter{}, nil
	}
	builder := newFilterBuilder(options, ratings)
	if err := readDataset(dir, reader.FileTitleBasics, builder.readBasics); err != nil {
		return titleFilter{}, fmt.Errorf("building filter: %w", err)
	}
	if err := readDataset(dir, reader.FileTitleEpisode, builder.readEpisodes); err != nil {
		return titleFilter{}, fmt.Errorf("building filter: %w", err)
	}
	return builder.filter(), nil
}

// readDataset opens one gzipped dataset file from dir and hands it to read,
// naming the file in any error, since it chose which one.
func readDataset(dir, file string, read func(io.Reader) error) error {
	f, err := reader.OpenGzip(filepath.Join(dir, file))
	if err != nil {
		return err
	}
	defer f.Close()
	if err := read(f); err != nil {
		return fmt.Errorf("%s: %w", file, err)
	}
	return nil
}

// filterBuilder accumulates the allow-list, one dataset file at a time.
type filterBuilder struct {
	options BuildOptions
	ratings map[int64]rating
	allowed *bitset.BitSet
}

func newFilterBuilder(options BuildOptions, ratings map[int64]rating) *filterBuilder {
	return &filterBuilder{options: options, ratings: ratings, allowed: bitset.New(0)}
}

// readBasics allows every title.basics record that all enabled rules accept.
func (b *filterBuilder) readBasics(basics io.Reader) error {
	var read int64
	for record, err := range reader.ReadTitleBasics(basics) {
		if err != nil {
			return err
		}
		read++
		id, err := titleID(record.Tconst)
		if err != nil {
			return rowError(read, record.Tconst, fmt.Errorf("tconst: %w", err))
		}
		if b.allowsTitle(record, id) {
			b.allowed.Set(id)
		}
	}
	return nil
}

// allowsTitle applies each enabled rule to one title, independently of the rest.
func (b *filterBuilder) allowsTitle(record reader.TitleBasics, id uint) bool {
	if b.options.NotAdult && record.IsAdult {
		return false
	}
	if b.options.Rated {
		if _, rated := b.ratings[int64(id)]; !rated {
			return false
		}
	}
	return true
}

// readEpisodes gives every episode the fate of its parent series, whatever rule
// decided that, which is what stops the rules thinning out a series.
func (b *filterBuilder) readEpisodes(episodes io.Reader) error {
	var read int64
	for record, err := range reader.ReadTitleEpisode(episodes) {
		if err != nil {
			return err
		}
		read++
		episode, err := titleID(record.Tconst)
		if err != nil {
			return rowError(read, record.Tconst, fmt.Errorf("tconst: %w", err))
		}
		parent, err := titleID(record.ParentTconst)
		if err != nil {
			return rowError(read, record.Tconst, fmt.Errorf("parentTconst: %w", err))
		}
		if b.allowed.Test(parent) {
			b.allowed.Set(episode)
			continue
		}
		b.allowed.Clear(episode)
	}
	return nil
}

// filter returns the finished allow-list.
func (b *filterBuilder) filter() titleFilter {
	return titleFilter{allowed: b.allowed}
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
