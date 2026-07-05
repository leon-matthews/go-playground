# Download, update, and query a local mirror of the Have I Been Pwned password database
#
# Requires:
#
# - revive (https://github.com/mgechev/revive)
# - govulncheck (https://go.dev/blog/vuln)
# - gotestsum (https://github.com/gotestyourself/gotestsum)



.PHONY: build run test lint vulncheck check clean

build:
	go build -o pwnedpasswords ./cmd

run:
	go run ./cmd

test:
	gotestsum --hide-summary=skipped -f testdox -- --failfast ./...

lint:
	revive ./...

vulncheck:
	govulncheck ./...

check: lint vulncheck

clean:
	rm -f pwnedpasswords *.pprof *.log
