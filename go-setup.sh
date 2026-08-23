#!/bin/bash
#
# Install the Go tools listed in go-setup.md
#
set -eu

PACKAGES=(
    go.etcd.io/bbolt/cmd/bbolt@latest
    golang.org/x/perf/cmd/benchstat@latest
    github.com/charmbracelet/glow@latest
    mvdan.cc/gofumpt@latest
    golang.org/x/tools/cmd/goimports@latest
    github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.13.1
    github.com/oligot/go-mod-upgrade@latest
    golang.org/x/tools/gopls@latest
    golang.org/x/vuln/cmd/govulncheck@latest
    github.com/Zxilly/go-size-analyzer/cmd/gsa@latest
    gotest.tools/gotestsum@latest
    github.com/nats-io/natscli/nats@latest
    go.uber.org/nilaway/cmd/nilaway@latest
    github.com/mgechev/revive@latest
    github.com/boyter/scc/v3@latest
    github.com/sqlc-dev/sqlc/cmd/sqlc@latest
    github.com/Antonboom/testifylint@latest
)

if ! command -v go > /dev/null; then
    echo "Error: 'go' not found in PATH" >&2
    exit 1
fi

for package in "${PACKAGES[@]}"; do
    echo "==> go install $package"
    go install "$package"
done

echo
echo "Installed ${#PACKAGES[@]} packages into $(go env GOPATH)/bin"

echo
echo "Clear all Go caches"
go clean -cache -testcache -modcache -fuzzcache
