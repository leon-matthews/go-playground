package mimicry

import (
	"cmp"
	"path/filepath"
	"slices"
)

// Summary is the total file count and combined size of a scan.
type Summary struct {
	Count int
	Size  int64
}

// ExtensionStat is the file count and combined size for a single extension.
type ExtensionStat struct {
	Extension string
	Count     int
	Size      int64
}

// DuplicateGroup is a set of two or more files sharing the same content hash.
type DuplicateGroup struct {
	Hash  [32]byte
	Size  int64
	Files []FileInfo
}

// Reclaimable is the space freed by keeping one copy of the group and deleting the rest.
func (g DuplicateGroup) Reclaimable() int64 {
	if len(g.Files) < 2 {
		return 0
	}
	return int64(len(g.Files)-1) * g.Size
}

// Summarize returns the file count and combined size of files.
func Summarize(files []FileInfo) Summary {
	var total int64
	for _, f := range files {
		total += f.Size
	}
	return Summary{Count: len(files), Size: total}
}

// ExtensionStats buckets files by extension, sorted by descending count then extension.
func ExtensionStats(files []FileInfo) []ExtensionStat {
	byExt := make(map[string]*ExtensionStat)
	for _, f := range files {
		s := byExt[f.Extension]
		if s == nil {
			s = &ExtensionStat{Extension: f.Extension}
			byExt[f.Extension] = s
		}
		s.Count++
		s.Size += f.Size
	}

	stats := make([]ExtensionStat, 0, len(byExt))
	for _, s := range byExt {
		stats = append(stats, *s)
	}
	slices.SortFunc(stats, func(a, b ExtensionStat) int {
		if a.Count != b.Count {
			return cmp.Compare(b.Count, a.Count)
		}
		return cmp.Compare(a.Extension, b.Extension)
	})
	return stats
}

// DuplicateGroups returns every set of two or more files sharing a content hash.
//
// Groups are ordered by reclaimable space descending; files within each group are ordered by
// path; files with a zero hash (never hashed) are excluded.
func DuplicateGroups(files []FileInfo) []DuplicateGroup {
	byHash := make(map[[32]byte][]FileInfo)
	for _, f := range files {
		if f.Hash == ([32]byte{}) {
			continue
		}
		byHash[f.Hash] = append(byHash[f.Hash], f)
	}

	var groups []DuplicateGroup
	for hash, members := range byHash {
		if len(members) < 2 {
			continue
		}
		slices.SortFunc(members, func(a, b FileInfo) int {
			return cmp.Compare(a.Path, b.Path)
		})
		groups = append(groups, DuplicateGroup{
			Hash:  hash,
			Size:  members[0].Size,
			Files: members,
		})
	}
	slices.SortFunc(groups, func(a, b DuplicateGroup) int {
		if ra, rb := a.Reclaimable(), b.Reclaimable(); ra != rb {
			return cmp.Compare(rb, ra)
		}
		return cmp.Compare(filepath.Base(a.Files[0].Path), filepath.Base(b.Files[0].Path))
	})
	return groups
}
