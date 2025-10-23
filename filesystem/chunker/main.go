package main

import (
	"crypto/sha256"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"runtime"
	"runtime/pprof"
	"slices"
	"sync"

	chunkers "github.com/PlakarKorp/go-cdc-chunkers"
	_ "github.com/PlakarKorp/go-cdc-chunkers/chunkers/fastcdc"
	_ "github.com/PlakarKorp/go-cdc-chunkers/chunkers/jc"
	_ "github.com/PlakarKorp/go-cdc-chunkers/chunkers/ultracdc"
)

var (
	numWorkers = runtime.NumCPU()
	queueDepth = numWorkers * 2
)

// Flags for pprof
var (
	cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")
	memprofile = flag.String("memprofile", "", "write memory profile to `file`")
)

func main() {
	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		defer f.Close() // error handling omitted for example
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}

	if len(flag.Args()) != 1 {
		log.Fatalf("Usage: %v PATH", path.Base(os.Args[0]))
	}

	log.Println("Starting with", numWorkers, "workers")
	reader, err := os.Open(flag.Arg(0))
	if err != nil {
		log.Fatal("Error opening file:", err)
	}

	chunks, err := chunkStream(reader)
	if err != nil {
		log.Fatal("Error creating stream", err)
	}

	hashed := hashStream(chunks)

	for c := range hashed {
		fmt.Printf("#%d %d %d %x\n", c.worker, c.offset, c.length, c.hash)
	}

	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			log.Fatal("could not create memory profile: ", err)
		}
		defer f.Close() // error handling omitted for example
		runtime.GC()    // get up-to-date statistics
		// Lookup("allocs") creates a profile similar to go test -memprofile.
		// Alternatively, use Lookup("heap") for a profile
		// that has inuse_space as the default index.
		if err := pprof.Lookup("allocs").WriteTo(f, 0); err != nil {
			log.Fatal("could not write memory profile: ", err)
		}
	}
}

type chunk struct {
	offset int
	data   []byte
	length int
	hash   []byte
	worker int
}

// hashStream populates the hash field of incoming chunks
func hashStream(input <-chan chunk) <-chan chunk {
	output := make(chan chunk, queueDepth)
	var wg sync.WaitGroup
	for id := range numWorkers {
		wg.Go(func() {
			hasher := sha256.New()
			for c := range input {
				hasher.Write(c.data)
				c.hash = hasher.Sum(nil)
				c.data = nil
				c.worker = id
				output <- c
				hasher.Reset()
			}
		})
	}

	go func() {
		wg.Wait()
		close(output)
	}()

	return output
}

// chunkStream streams chunks of data from reader with the data, but not hash field, populated.
func chunkStream(reader io.Reader) (<-chan chunk, error) {
	output := make(chan chunk, queueDepth)
	options := &chunkers.ChunkerOpts{
		MinSize:    0x01 << 19, // 512KiB
		MaxSize:    0x01 << 22, // 4Mib
		NormalSize: 0x01 << 20, // 1MiB
	}
	chunker, err := chunkers.NewChunker("jc-v1.0.0", reader, options)
	if err != nil {
		return nil, fmt.Errorf("initialising CDC chunker: %w", err)
	}
	offset := 0

	go func() {
		for {
			c, err := chunker.Next()
			if err != nil && err != io.EOF {
				panic(fmt.Sprintf("Fatal error calling chunker.Next():  %s"))
			}
			// Preserve slice even if (when!) underlying array changes
			c = slices.Clone(c)
			numBytes := len(c)
			output <- chunk{offset: offset, data: c, length: numBytes}
			if err == io.EOF {
				break
			}
			offset += numBytes
		}
		close(output)
	}()

	return output, nil
}
