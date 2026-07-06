# Concurrency levels for the benchmark sweep; override with e.g. make bench CPUS=1,2,4
CPUS ?= 1,2,4,8,16,32,64

.PHONY: bench
bench:
	go test -run '^$$' -bench . -benchmem -cpu $(CPUS)
