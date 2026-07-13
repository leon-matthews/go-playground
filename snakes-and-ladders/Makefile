# Linting requires:
# - revive (https://github.com/mgechev/revive)
# - govulncheck (https://go.dev/blog/vuln)
BINARY := snakes-and-ladders
CMD := ./cmd/go_ladders

.PHONY: build run test lint vulncheck check clean default.pgo

build:
	go build -o $(BINARY) $(CMD)

run:
	go run $(CMD)

# Regenerate the PGO profile then copy it into where go build expects it
default.pgo:
	go build -o $(BINARY) $(CMD)
	./$(BINARY) --profile
	mv default.pgo $(CMD)/default.pgo
	rm -f $(BINARY)

test:
	go test ./...

lint:
	revive ./...

vulncheck:
	govulncheck ./...

check: lint test vulncheck

clean:
	rm -f $(BINARY) *.pprof *.log default.pgo
