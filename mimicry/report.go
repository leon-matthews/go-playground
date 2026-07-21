package mimicry

import (
	"path/filepath"
	"sort"
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
	sort.Slice(stats, func(i, j int) bool {
		if stats[i].Count != stats[j].Count {
			return stats[i].Count > stats[j].Count
		}
		return stats[i].Extension < stats[j].Extension
	})
	return stats
}

// DuplicateGroups returns every set of two or more files sharing a content hash, largest first.
//
// Files within each group are ordered by path; files with a zero hash (never hashed) are excluded.
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
		sort.Slice(members, func(i, j int) bool {
			return members[i].Path < members[j].Path
		})
		groups = append(groups, DuplicateGroup{
			Hash:  hash,
			Size:  members[0].Size,
			Files: members,
		})
	}
	sort.Slice(groups, func(i, j int) bool {
		if groups[i].Size != groups[j].Size {
			return groups[i].Size > groups[j].Size
		}
		return filepath.Base(groups[i].Files[0].Path) < filepath.Base(groups[j].Files[0].Path)
	})
	return groups
}
