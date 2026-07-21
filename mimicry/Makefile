# Run, test, and lint the mimicry duplicate-file scanner
#
# Requires:
#
# - revive (https://github.com/mgechev/revive)
# - gotestsum (https://github.com/gotestyourself/gotestsum)



.PHONY: build run test lint clean

build:
	go build -o mimicry ./cmd

run:
	go run ./cmd

test:
	gotestsum --hide-summary=skipped -f testdox -- --failfast ./...

lint:
	revive ./...

clean:
	rm -f mimicry *.pprof *.log
