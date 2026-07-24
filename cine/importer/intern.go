package importer

import (
	"cmp"
	"fmt"
	"slices"
	"strconv"
)

// maxBit is the highest bit position a bitmask column can hold in a signed
// 64-bit integer.
const maxBit = 62

// interner assigns a small integer id to each distinct string it sees, in
// first-seen order.
//
// For a lookup foreign key the id is the value stored in the row; for a bitmask
// column the id is the bit position. Queries resolve values by name through the
// lookup table, so the exact ids need not be stable across builds.
type interner struct {
	ids map[string]int64
}

func newInterner() *interner {
	return &interner{ids: make(map[string]int64)}
}

// id returns the id for name, assigning the next unused one on first sight.
func (in *interner) id(name string) int64 {
	if id, ok := in.ids[name]; ok {
		return id
	}
	id := int64(len(in.ids))
	in.ids[name] = id
	return id
}

// bit returns the bitmask contribution for name: 1 shifted left by its id.
// It panics if a column accumulates more distinct values than a 64-bit mask
// holds, which signals that the field has outgrown a single-integer bitmask.
func (in *interner) bit(name string) int64 {
	id := in.id(name)
	if id > maxBit {
		panic(fmt.Sprintf("importer: bitmask column exceeded %d values at %q", maxBit+1, name))
	}
	return 1 << id
}

// lookupEntry is one id/name pair destined for a lookup table.
type lookupEntry struct {
	id   int64
	name string
}

// entries returns the interned pairs ordered by id, ready for the lookup table.
func (in *interner) entries() []lookupEntry {
	entries := make([]lookupEntry, 0, len(in.ids))
	for name, id := range in.ids {
		entries = append(entries, lookupEntry{id: id, name: name})
	}
	slices.SortFunc(entries, func(a, b lookupEntry) int { return cmp.Compare(a.id, b.id) })
	return entries
}

// parseID turns an IMDb identifier such as "tt0133093" or "nm0000001" into the
// integer that remains once its two-letter prefix is dropped.
func parseID(s string) (int64, error) {
	if len(s) < 3 {
		return 0, fmt.Errorf("invalid identifier %q", s)
	}
	n, err := strconv.ParseInt(s[2:], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid identifier %q: %w", s, err)
	}
	return n, nil
}
