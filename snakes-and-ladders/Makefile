# Linting requires:
# - revive (https://github.com/mgechev/revive)
# - govulncheck (https://go.dev/blog/vuln)
BINARY := snakes-and-ladders

.PHONY: build run test lint vulncheck check clean

build:
	go build -o $(BINARY) .

run:
	go run .

test:
	go test ./...

lint:
	revive ./...

vulncheck:
	govulncheck ./...

check: lint test vulncheck

clean:
	rm -f $(BINARY) *.pprof *.log
