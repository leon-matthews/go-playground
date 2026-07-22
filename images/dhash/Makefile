# Requires:
# - revive (https://github.com/mgechev/revive)
# - gotestsum (https://github.com/gotestyourself/gotestsum)

BINARY := dhash

.PHONY: build run test lint clean

build:
	go build -o $(BINARY) ./cmd

run:
	go run ./cmd

test:
	gotestsum --hide-summary=skipped -f testdox -- --failfast ./...

lint:
	revive ./...

clean:
	rm -f $(BINARY) *.pprof *.log
