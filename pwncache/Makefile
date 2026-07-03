# Download, update, and query a local mirror of the Have I Been Pwned password database
#
# Requires:
#
# - revive (https://github.com/mgechev/revive)
# - govulncheck (https://go.dev/blog/vuln)
.PHONY: build run test lint vulncheck check clean

build:
	go build -o pwncache ./cmd

run:
	go run ./cmd

test:
	go test ./...

lint:
	revive ./...

vulncheck:
	govulncheck ./...

check: lint test vulncheck

clean:
	rm -f pwncache *.pprof *.log
