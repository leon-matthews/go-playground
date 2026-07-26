package importer

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/bits-and-blooms/bitset"

	"local.dev/cine/reader"
)

// FilterRules selects which row filters a build applies.
//
// The same value drives the filter and is recorded in build_info, so a database
// cannot claim a rule that did not run. The zero value filters nothing.
type FilterRules struct {
	Rated    bool // keep only titles IMDb has published a rating for
	NotAdult bool // drop titles flagged isAdult
}

// any reports whether any rule is enabled, and so whether a filter is built.
func (r FilterRules) any() bool {
	return r.Rated || r.NotAdult
}

// String names the enabled rules for a log line, in build_info's column order.
func (r FilterRules) String() string {
	var names []string
	if r.Rated {
		names = append(names, "rated")
	}
	if r.NotAdult {
		names = append(names, "not-adult")
	}
	if len(names) == 0 {
		return "none"
	}
	return strings.Join(names, ",")
}

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
func buildFilter(dir string, rules FilterRules, ratings map[int64]rating) (titleFilter, error) {
	if !rules.any() {
		return titleFilter{}, nil
	}
	builder := newFilterBuilder(rules, ratings)
	if err := readDataset(dir, reader.FileTitleBasics, builder.readBasics); err != nil {
		return titleFilter{}, err
	}
	if err := readDataset(dir, reader.FileTitleEpisode, builder.readEpisodes); err != nil {
		return titleFilter{}, err
	}
	return builder.filter(), nil
}

// readDataset opens one gzipped dataset file from dir and hands it to read.
func readDataset(dir, file string, read func(io.Reader) error) error {
	f, err := reader.OpenGzip(filepath.Join(dir, file))
	if err != nil {
		return err
	}
	defer f.Close()
	return read(f)
}

// filterBuilder accumulates the allow-list, one dataset file at a time.
type filterBuilder struct {
	rules   FilterRules
	ratings map[int64]rating
	allowed *bitset.BitSet
}

func newFilterBuilder(rules FilterRules, ratings map[int64]rating) *filterBuilder {
	return &filterBuilder{rules: rules, ratings: ratings, allowed: bitset.New(0)}
}

// readBasics allows every title.basics record that all enabled rules accept.
func (b *filterBuilder) readBasics(basics io.Reader) error {
	for record, err := range reader.ReadTitleBasics(basics) {
		if err != nil {
			return err
		}
		id, err := titleID(record.Tconst)
		if err != nil {
			return err
		}
		if b.allowsTitle(record, id) {
			b.allowed.Set(id)
		}
	}
	return nil
}

// allowsTitle applies each enabled rule to one title, independently of the rest.
func (b *filterBuilder) allowsTitle(record reader.TitleBasics, id uint) bool {
	if b.rules.NotAdult && record.IsAdult {
		return false
	}
	if b.rules.Rated {
		if _, rated := b.ratings[int64(id)]; !rated {
			return false
		}
	}
	return true
}

// readEpisodes gives every episode the fate of its parent series, whatever rule
// decided that, so no series is ever stored with gaps in it.
func (b *filterBuilder) readEpisodes(episodes io.Reader) error {
	for record, err := range reader.ReadTitleEpisode(episodes) {
		if err != nil {
			return err
		}
		episode, err := titleID(record.Tconst)
		if err != nil {
			return err
		}
		parent, err := titleID(record.ParentTconst)
		if err != nil {
			return err
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
