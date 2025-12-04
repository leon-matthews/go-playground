
# Go Performance Measurement

1. pprof: Batched or live profile data (CPU, memory, block, mutex).
2. trace: After-the-fact timeline of events: goroutine scheduling, system calls, etc...

## Minimal pprof

For real applications, use pprof to identify allocation hotspots:

import (
    "os"
    "runtime/pprof"
)

func main() {
    f, _ := os.Create("heap.prof")
    defer f.Close()
    
    // ... run your application
    
    pprof.WriteHeapProfile(f)
}
Then analyze:

go tool pprof -http=:8080 heap.prof
This opens an interactive visualization showing exactly where your allocations come from.

For benchmark-driven profiling:

go test -bench=. -memprofile=mem.prof
go tool pprof -alloc_space mem.prof

