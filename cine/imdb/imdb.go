// Package imdb reads the IMDb non-commercial dataset TSV files into typed
// Go records.
//
// Each dataset file has a Read function that streams its rows from an
// io.Reader as an iter.Seq2 of record and error. The records are a faithful
// view of the file: identifiers keep their "tt"/"nm" string form and IMDb's
// \N null marker becomes a nil pointer or an empty value.
//
// See https://developer.imdb.com/non-commercial-datasets/ for the dataset.
package imdb

import (
	"bufio"
	"fmt"
	"io"
	"iter"
	"os"
	"strings"

	"github.com/klauspost/compress/gzip"
)

// Canonical file names of the seven dataset files.
const (
	FileNameBasics      = "name.basics.tsv.gz"
	FileTitleAkas       = "title.akas.tsv.gz"
	FileTitleBasics     = "title.basics.tsv.gz"
	FileTitleCrew       = "title.crew.tsv.gz"
	FileTitleEpisode    = "title.episode.tsv.gz"
	FileTitlePrincipals = "title.principals.tsv.gz"
	FileTitleRatings    = "title.ratings.tsv.gz"
)

// maxLineBytes caps a single TSV row. Real rows sit far below this; the limit
// only stops a corrupt stream with no newline from exhausting memory.
const maxLineBytes = 1 << 20

// read streams typed records from an IMDb TSV stream.
//
// The first line must be the header, whose columns are checked against want so
// a change to the dataset layout fails loudly rather than misparsing silently.
// Each later line is split on tabs and handed to fromFields. A row with the
// wrong column count, or one fromFields rejects, yields an error carrying the
// line number without ending the iteration.
func read[T any](r io.Reader, want []string, fromFields func([]string) (T, error)) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		var zero T
		scanner := bufio.NewScanner(r)
		scanner.Buffer(make([]byte, 0, 64*1024), maxLineBytes)

		// An empty stream is simply no records; only a header row is validated
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				yield(zero, fmt.Errorf("reading header: %w", err))
			}
			return
		}
		if err := checkHeader(strings.Split(scanner.Text(), "\t"), want); err != nil {
			yield(zero, err)
			return
		}

		for line := 2; scanner.Scan(); line++ {
			fields := strings.Split(scanner.Text(), "\t")
			if len(fields) != len(want) {
				if !yield(zero, fmt.Errorf("line %d: got %d fields, want %d", line, len(fields), len(want))) {
					return
				}
				continue
			}
			record, err := fromFields(fields)
			if err != nil {
				if !yield(zero, fmt.Errorf("line %d: %w", line, err)) {
					return
				}
				continue
			}
			if !yield(record, nil) {
				return
			}
		}
		if err := scanner.Err(); err != nil {
			yield(zero, fmt.Errorf("reading rows: %w", err))
		}
	}
}

// checkHeader verifies a file's header row matches the columns want.
func checkHeader(got, want []string) error {
	if len(got) != len(want) {
		return fmt.Errorf("header has %d columns, want %d: %v", len(got), len(want), want)
	}
	for i := range want {
		if got[i] != want[i] {
			return fmt.Errorf("header column %d is %q, want %q", i+1, got[i], want[i])
		}
	}
	return nil
}

// OpenTSV opens a gzipped IMDb dataset file for reading.
//
// Close the returned reader to release both the gzip stream and the file.
func OpenTSV(path string) (io.ReadCloser, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	gz, err := gzip.NewReader(file)
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("opening gzip %s: %w", path, err)
	}
	return gzipFile{gz: gz, file: file}, nil
}

// gzipFile couples a gzip reader with its backing file so both close together.
type gzipFile struct {
	gz   *gzip.Reader
	file *os.File
}

func (g gzipFile) Read(p []byte) (int, error) { return g.gz.Read(p) }

func (g gzipFile) Close() error {
	err := g.gz.Close()
	if ferr := g.file.Close(); err == nil {
		err = ferr
	}
	return err
}
