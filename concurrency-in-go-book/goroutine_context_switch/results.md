
# Measure Context Switching Overhead

Linux 6.14 and AMD Ryzen 9 5950X (16-cores, 32-threads)

## Native threads

```
$ taskset -c 0 perf bench sched pipe -T
# Running 'sched/pipe' benchmark:
# Executed 1000000 pipe operations between two threads
    Total time: 5.315 [sec]
    5.315496 usecs/op
    188129 ops/sec
```

5.3us to send and receive message on a thread is 2.6us per context switch

## Goroutines

```
$ go test -bench=. --cpu=1 overhead_test.go
goos: linux
goarch: amd64
cpu: AMD Ryzen 9 5950X 16-Core Processor
BenchmarkContextSwitch 	12363787	        93.34 ns/op
PASS
ok  	command-line-arguments	1.255s
```

94ns per context switch! That is **28 times faster** than native threads.
