// Command reader-test reads every record from every IMDb dataset file under a
// folder, under a CPU profile, to validate and optimise the reader package.
//
// Usage:
//
//	reader-test <imdb-data-folder>
//
// It logs any error encountered while reading, prints a per-file and total
// summary of records, errors, and throughput, and writes cpu.pprof to the
// working directory.
package main

import (
	"fmt"
	"io"
	"iter"
	"log"
	"os"
	"path/filepath"
	"runtime/pprof"
	"time"

	"local.dev/cine/reader"
)

// summaryLine formats one row of the records/errors/throughput report.
const summaryLine = "%-23s %10d records  %6d errors  %10s  %10.0f rec/s\n"

func main() {
	log.SetFlags(0)
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "usage: %s <imdb-data-folder>\n", filepath.Base(os.Args[0]))
		os.Exit(2)
	}
	if err := run(os.Args[1]); err != nil {
		log.Fatal(err)
	}
}

// run reads every dataset file in folder under a CPU profile.
func run(folder string) error {
	info, err := os.Stat(folder)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", folder)
	}

	stop, err := startCPUProfile("cpu.pprof")
	if err != nil {
		return err
	}
	defer stop()

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
		records, errors := read(folder)
		totalRecords += records
		totalErrors += errors
	}
	report("TOTAL", totalRecords, totalErrors, time.Since(start))
	return nil
}

// fileReader reads every record of one dataset file, printing a summary line
// and returning the record and error counts.
type fileReader func(folder string) (records, errors int)

// readFile builds a fileReader for the named file and its typed reader.
func readFile[T any](name string, readRecords func(io.Reader) iter.Seq2[T, error]) fileReader {
	return func(folder string) (records, errors int) {
		file, err := reader.OpenTSV(filepath.Join(folder, name))
		if err != nil {
			log.Printf("%s: %v", name, err)
			return 0, 1
		}
		defer file.Close()

		start := time.Now()
		for _, err := range readRecords(file) {
			if err != nil {
				log.Printf("%s: %v", name, err)
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
	fmt.Printf(summaryLine, name, records, errors, elapsed.Round(time.Millisecond), rate(records, elapsed))
}

// rate returns records per second, or zero when no time has elapsed.
func rate(records int, elapsed time.Duration) float64 {
	if elapsed <= 0 {
		return 0
	}
	return float64(records) / elapsed.Seconds()
}

// startCPUProfile begins CPU profiling to path; the returned function stops the
// profile and closes the file.
func startCPUProfile(path string) (func(), error) {
	file, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("creating cpu profile: %w", err)
	}
	if err := pprof.StartCPUProfile(file); err != nil {
		file.Close()
		return nil, fmt.Errorf("starting cpu profile: %w", err)
	}
	return func() {
		pprof.StopCPUProfile()
		file.Close()
	}, nil
}
