package main

import (
	"fmt"
	"io"
	"iter"
	"path/filepath"
	"time"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"

	"local.dev/cine/common"
	"local.dev/cine/reader"
)

// summaryLine formats one row of the records/errors/throughput report.
const summaryLine = "%-23s %13s records  %7s errors  %10s  %13s rec/s\n"

// newReaderBenchmarkCmd builds the "reader-benchmark" sub-command.
func newReaderBenchmarkCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "reader-benchmark <imdb-data-folder>",
		Short: "Read every dataset record, reporting throughput",
		Long: `Read every dataset record, reporting throughput.

Reads every record from every IMDb dataset file in the folder to validate and
optimise the reader package. Any error encountered is logged, and a per-file and
total summary of records, errors and throughput is printed. Pass --profile to
write a CPU profile of the run.`,
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return benchmarkReaders(args[0], newLogger())
		},
	}
}

// benchmarkReaders reads every dataset file in folder, printing a summary of each.
func benchmarkReaders(folder string, logger *log.Logger) error {
	if err := requireFolder(folder); err != nil {
		return err
	}

	files := []fileReader{
		readFile(reader.FileNameBasics, reader.ReadNameBasics),
		readFile(reader.FileTitleAkas, reader.ReadTitleAkas),
		readFile(reader.FileTitleBasics, reader.ReadTitleBasics),
		readFile(reader.FileTitleCrew, reader.ReadTitleCrew),
		readFile(reader.FileTitleEpisode, reader.ReadTitleEpisode),
		readFile(reader.FileTitlePrincipals, reader.ReadTitlePrincipals),
		readFile(reader.FileTitleRatings, reader.ReadTitleRatings),
	}

	start := time.Now()
	var totalRecords, totalErrors int
	for _, read := range files {
		records, errors := read(folder, logger)
		totalRecords += records
		totalErrors += errors
	}
	report("TOTAL", totalRecords, totalErrors, time.Since(start))
	return nil
}

// fileReader reads every record of one dataset file, printing a summary line
// and returning the record and error counts.
type fileReader func(folder string, logger *log.Logger) (records, errors int)

// readFile builds a fileReader for the named file and its typed reader.
func readFile[T any](name string, readRecords func(io.Reader) iter.Seq2[T, error]) fileReader {
	return func(folder string, logger *log.Logger) (records, errors int) {
		file, err := reader.OpenGzip(filepath.Join(folder, name))
		if err != nil {
			logger.Error("could not open file", "file", name, "err", err)
			return 0, 1
		}
		defer file.Close()

		start := time.Now()
		for _, err := range readRecords(file) {
			if err != nil {
				logger.Error("could not read record", "file", name, "err", err)
				errors++
				continue
			}
			records++
		}
		report(name, records, errors, time.Since(start))
		return records, errors
	}
}

// report prints one summary line with the throughput derived from elapsed.
func report(name string, records, errors int, elapsed time.Duration) {
	fmt.Printf(
		summaryLine,
		name,
		common.Commas(int64(records)),
		common.Commas(int64(errors)),
		elapsed.Round(time.Millisecond),
		common.Commas(rate(records, elapsed)),
	)
}

// rate returns records per second, or zero when no time has elapsed.
func rate(records int, elapsed time.Duration) int64 {
	if elapsed <= 0 {
		return 0
	}
	return int64(float64(records) / elapsed.Seconds())
}
