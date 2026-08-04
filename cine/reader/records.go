package reader

import (
	"fmt"
	"io"
	"iter"
)

// NameBasics is one row of name.basics.tsv: a person's core biography.
type NameBasics struct {
	Nconst            string   // "nm0000001"
	PrimaryName       string   // "Fred Astaire"; "" for \N
	BirthYear         int      // Missing (-1) for \N; map to SQL NULL
	DeathYear         int      // Missing (-1) for \N; map to SQL NULL
	PrimaryProfession []string // "actor,producer" -> ["actor", "producer"]
	KnownForTitles    []string // tconsts the person is known for
}

var nameBasicsHeader = []string{
	"nconst", "primaryName", "birthYear", "deathYear",
	"primaryProfession", "knownForTitles",
}

// ReadNameBasics streams the rows of a name.basics TSV stream.
func ReadNameBasics(r io.Reader) iter.Seq2[NameBasics, error] {
	return read(r, nameBasicsHeader, func(c *cursor) (NameBasics, error) {
		return NameBasics{
			Nconst:            c.str(0),
			PrimaryName:       c.optionalStr(1),
			BirthYear:         c.optionalInt(2),
			DeathYear:         c.optionalInt(3),
			PrimaryProfession: c.list(4),
			KnownForTitles:    c.list(5),
		}, c.err
	})
}

// TitleAkas is one row of title.akas.tsv: a localized or alternate title.
type TitleAkas struct {
	TitleID         string   // "tt0000001"
	Ordering        int      // 1
	Title           string   // localized title
	Region          string   // "" for \N
	Language        string   // "" for \N
	Types           []string // \x02-separated list, eg. ["imdbDisplay", "dvd"]
	Attributes      []string // \x02-separated list, eg. ["literal title"]
	IsOriginalTitle bool     // 1 -> true
}

var titleAkasHeader = []string{
	"titleId", "ordering", "title", "region",
	"language", "types", "attributes", "isOriginalTitle",
}

// ReadTitleAkas streams the rows of a title.akas TSV stream.
func ReadTitleAkas(r io.Reader) iter.Seq2[TitleAkas, error] {
	return read(r, titleAkasHeader, func(c *cursor) (TitleAkas, error) {
		return TitleAkas{
			TitleID:         c.str(0),
			Ordering:        c.requiredInt(1),
			Title:           c.str(2),
			Region:          c.optionalStr(3),
			Language:        c.optionalStr(4),
			Types:           c.akasArray(5),
			Attributes:      c.akasArray(6),
			IsOriginalTitle: c.boolean(7),
		}, c.err
	})
}

// TitleBasics is one row of title.basics.tsv: a title's core details.
type TitleBasics struct {
	Tconst         string   // "tt0000001"
	TitleType      string   // "short"
	PrimaryTitle   string   // promotional title
	OriginalTitle  string   // title in the original language
	IsAdult        bool     // 1 -> true
	StartYear      int      // Missing (-1) for \N; map to SQL NULL
	EndYear        int      // Missing (-1) for \N; map to SQL NULL
	RuntimeMinutes int      // Missing (-1) for \N; map to SQL NULL
	Genres         []string // "Documentary,Short" -> ["Documentary", "Short"]
}

var titleBasicsHeader = []string{
	"tconst", "titleType", "primaryTitle", "originalTitle", "isAdult",
	"startYear", "endYear", "runtimeMinutes", "genres",
}

// ReadTitleBasics streams the rows of a title.basics TSV stream.
func ReadTitleBasics(r io.Reader) iter.Seq2[TitleBasics, error] {
	return read(r, titleBasicsHeader, func(c *cursor) (TitleBasics, error) {
		return TitleBasics{
			Tconst:         c.str(0),
			TitleType:      c.str(1),
			PrimaryTitle:   c.str(2),
			OriginalTitle:  c.str(3),
			IsAdult:        c.boolean(4),
			StartYear:      c.optionalInt(5),
			EndYear:        c.optionalInt(6),
			RuntimeMinutes: c.optionalInt(7),
			Genres:         c.list(8),
		}, c.err
	})
}

// TitleCrew is one row of title.crew.tsv: a title's directors and writers.
type TitleCrew struct {
	Tconst    string   // "tt0000001"
	Directors []string // director nconsts
	Writers   []string // writer nconsts
}

var titleCrewHeader = []string{"tconst", "directors", "writers"}

// ReadTitleCrew streams the rows of a title.crew TSV stream.
func ReadTitleCrew(r io.Reader) iter.Seq2[TitleCrew, error] {
	return read(r, titleCrewHeader, func(c *cursor) (TitleCrew, error) {
		return TitleCrew{
			Tconst:    c.str(0),
			Directors: c.list(1),
			Writers:   c.list(2),
		}, c.err
	})
}

// TitleEpisode is one row of title.episode.tsv: an episode's place in a series.
type TitleEpisode struct {
	Tconst        string // episode tconst
	ParentTconst  string // series tconst
	SeasonNumber  int    // Missing (-1) for \N; map to SQL NULL
	EpisodeNumber int    // Missing (-1) for \N; map to SQL NULL
}

// HasParent reports whether IMDb named a parent series for this episode.
//
// The dataset documents parentTconst as required, so it is read as a plain
// string, but some exports have had a rare \N instead of a tconst.
func (e TitleEpisode) HasParent() bool {
	return e.ParentTconst != nullMarker
}

var titleEpisodeHeader = []string{"tconst", "parentTconst", "seasonNumber", "episodeNumber"}

// ReadTitleEpisode streams the rows of a title.episode TSV stream.
func ReadTitleEpisode(r io.Reader) iter.Seq2[TitleEpisode, error] {
	return read(r, titleEpisodeHeader, func(c *cursor) (TitleEpisode, error) {
		return TitleEpisode{
			Tconst:        c.str(0),
			ParentTconst:  c.str(1),
			SeasonNumber:  c.optionalInt(2),
			EpisodeNumber: c.optionalInt(3),
		}, c.err
	})
}

// TitlePrincipals is one row of title.principals.tsv: a title's cast or crew credit.
type TitlePrincipals struct {
	Tconst     string   // "tt0000001"
	Ordering   int      // 1
	Nconst     string   // person nconst
	Category   string   // "actor", "director", ...
	Job        string   // specific job, "" for \N
	Characters []string // Weirdly, JSON eg. ["Self"] -> ["Self"], nil for \N
}

var titlePrincipalsHeader = []string{"tconst", "ordering", "nconst", "category", "job", "characters"}

// ReadTitlePrincipals streams the rows of a title.principals TSV stream.
func ReadTitlePrincipals(r io.Reader) iter.Seq2[TitlePrincipals, error] {
	return read(r, titlePrincipalsHeader, func(c *cursor) (TitlePrincipals, error) {
		return TitlePrincipals{
			Tconst:     c.str(0),
			Ordering:   c.requiredInt(1),
			Nconst:     c.str(2),
			Category:   c.str(3),
			Job:        c.optionalStr(4),
			Characters: c.characters(5),
		}, c.err
	})
}

// TitleRatings is one row of title.ratings.tsv: a title's aggregate rating.
type TitleRatings struct {
	Tconst        string  // "tt0000001"
	AverageRating float64 // 5.7
	NumVotes      int     // 2220
}

var titleRatingsHeader = []string{"tconst", "averageRating", "numVotes"}

// ReadTitleRatings streams the rows of a title.ratings TSV stream.
func ReadTitleRatings(r io.Reader) iter.Seq2[TitleRatings, error] {
	return read(r, titleRatingsHeader, func(c *cursor) (TitleRatings, error) {
		return TitleRatings{
			Tconst:        c.str(0),
			AverageRating: c.float(1),
			NumVotes:      c.requiredInt(2),
		}, c.err
	})
}

// cursor extracts typed values from one row's tab-separated fields.
//
// Captures the first parse error encountered, named for the column it came from:
// the field index alone leaves the caller counting columns in a file of nine.
type cursor struct {
	fields  []string
	columns []string // the file's header, so an error can name its column
	err     error
}

func (c *cursor) str(i int) string {
	return c.fields[i]
}

func (c *cursor) optionalStr(i int) string {
	return optionalString(c.fields[i])
}

func (c *cursor) list(i int) []string {
	return splitList(c.fields[i])
}

func (c *cursor) akasArray(i int) []string {
	return splitAkasArray(c.fields[i])
}

func (c *cursor) optionalInt(i int) int {
	if c.err != nil {
		return Missing
	}
	n, err := optionalInt(c.fields[i])
	c.keep(i, err)
	return n
}

func (c *cursor) requiredInt(i int) int {
	if c.err != nil {
		return 0
	}
	n, err := requiredInt(c.fields[i])
	c.keep(i, err)
	return n
}

func (c *cursor) boolean(i int) bool {
	if c.err != nil {
		return false
	}
	b, err := parseBool(c.fields[i])
	c.keep(i, err)
	return b
}

func (c *cursor) float(i int) float64 {
	if c.err != nil {
		return 0
	}
	f, err := parseFloat(c.fields[i])
	c.keep(i, err)
	return f
}

func (c *cursor) characters(i int) []string {
	if c.err != nil {
		return nil
	}
	names, err := parseCharacters(c.fields[i])
	c.keep(i, err)
	return names
}

// keep records the first error seen, naming the column it came from.
func (c *cursor) keep(i int, err error) {
	if err != nil && c.err == nil {
		c.err = fmt.Errorf("%s: %w", c.columns[i], err)
	}
}
