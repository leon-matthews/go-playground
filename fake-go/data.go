package fake

import (
	"embed"
	"strings"
	"sync"
)

// dataFS holds the word lists embedded at build time.
//
//go:embed data/*.txt
var dataFS embed.FS

// Shared word lists, each loaded from the embedded files on first use.
var (
	linesCities           = sync.OnceValue(func() []string { return loadLines("cities.txt") })
	linesFirstNames       = sync.OnceValue(func() []string { return loadLines("first_names_male.txt", "first_names_female.txt") })
	linesFirstNamesFemale = sync.OnceValue(func() []string { return loadLines("first_names_female.txt") })
	linesFirstNamesMale   = sync.OnceValue(func() []string { return loadLines("first_names_male.txt") })
	linesJobs             = sync.OnceValue(func() []string { return loadLines("jobs.txt") })
	linesLastNames        = sync.OnceValue(func() []string { return loadLines("last_names.txt") })
	linesStreets          = sync.OnceValue(func() []string { return loadLines("streets.txt") })
	linesSuburbs          = sync.OnceValue(func() []string { return loadLines("suburbs.txt") })
)

// loadLines reads the named data files, concatenating their lines.
//
// Blank lines and lines beginning with '#' are skipped, matching the Python
// load_lines helper. It panics if a file is missing, since the files are
// embedded and their absence is a build-time error.
func loadLines(names ...string) []string {
	var lines []string
	for _, name := range names {
		content, err := dataFS.ReadFile("data/" + name)
		if err != nil {
			panic("fake: missing data file: " + name)
		}
		for line := range strings.SplitSeq(string(content), "\n") {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			lines = append(lines, line)
		}
	}
	return lines
}

// choose returns a random line from the given list.
func (f *Faker) choose(lines []string) string {
	return lines[f.rng.IntN(len(lines))]
}
