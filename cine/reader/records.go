package reader

import (
	"io"
	"iter"
)

// NameBasics is one row of name.basics.tsv: a person's core biography.
type NameBasics struct {
	Nconst            string   // "nm0000001"
	PrimaryName       string   // "Fred Astaire"
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
	return read(r, nameBasicsHeader, func(f []string) (NameBasics, error) {
		c := cursor{fields: f}
		return NameBasics{
			Nconst:            c.str(0),
			PrimaryName:       c.str(1),
			BirthYear:         c.optInt(2),
			DeathYear:         c.optInt(3),
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
	Types           []string // "imdbDisplay" -> ["imdbDisplay"]
	Attributes      []string // "literal title" -> ["literal title"]
	IsOriginalTitle bool     // 1 -> true
}

var titleAkasHeader = []string{
	"titleId", "ordering", "title", "region",
	"language", "types", "attributes", "isOriginalTitle",
}

// ReadTitleAkas streams the rows of a title.akas TSV stream.
func ReadTitleAkas(r io.Reader) iter.Seq2[TitleAkas, error] {
	return read(r, titleAkasHeader, func(f []string) (TitleAkas, error) {
		c := cursor{fields: f}
		return TitleAkas{
			TitleID:         c.str(0),
			Ordering:        c.reqInt(1),
			Title:           c.str(2),
			Region:          c.optStr(3),
			Language:        c.optStr(4),
			Types:           c.list(5),
			Attributes:      c.list(6),
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
	return read(r, titleBasicsHeader, func(f []string) (TitleBasics, error) {
		c := cursor{fields: f}
		return TitleBasics{
			Tconst:         c.str(0),
			TitleType:      c.str(1),
			PrimaryTitle:   c.str(2),
			OriginalTitle:  c.str(3),
			IsAdult:        c.boolean(4),
			StartYear:      c.optInt(5),
			EndYear:        c.optInt(6),
			RuntimeMinutes: c.optInt(7),
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
	return read(r, titleCrewHeader, func(f []string) (TitleCrew, error) {
		c := cursor{fields: f}
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

var titleEpisodeHeader = []string{"tconst", "parentTconst", "seasonNumber", "episodeNumber"}

// ReadTitleEpisode streams the rows of a title.episode TSV stream.
func ReadTitleEpisode(r io.Reader) iter.Seq2[TitleEpisode, error] {
	return read(r, titleEpisodeHeader, func(f []string) (TitleEpisode, error) {
		c := cursor{fields: f}
		return TitleEpisode{
			Tconst:        c.str(0),
			ParentTconst:  c.str(1),
			SeasonNumber:  c.optInt(2),
			EpisodeNumber: c.optInt(3),
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
	Characters []string // JSON ["Self"] -> ["Self"], nil for \N
}

var titlePrincipalsHeader = []string{"tconst", "ordering", "nconst", "category", "job", "characters"}

// ReadTitlePrincipals streams the rows of a title.principals TSV stream.
func ReadTitlePrincipals(r io.Reader) iter.Seq2[TitlePrincipals, error] {
	return read(r, titlePrincipalsHeader, func(f []string) (TitlePrincipals, error) {
		c := cursor{fields: f}
		return TitlePrincipals{
			Tconst:     c.str(0),
			Ordering:   c.reqInt(1),
			Nconst:     c.str(2),
			Category:   c.str(3),
			Job:        c.optStr(4),
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
	return read(r, titleRatingsHeader, func(f []string) (TitleRatings, error) {
		c := cursor{fields: f}
		return TitleRatings{
			Tconst:        c.str(0),
			AverageRating: c.float(1),
			NumVotes:      c.reqInt(2),
		}, c.err
	})
}
