#!/bin/bash
# Install the Go tools listed in go-setup.md
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
    github.com/tsliwowicz/go-wrk@latest
    go.uber.org/nilaway/cmd/nilaway@latest
    github.com/mgechev/revive@latest
    github.com/boyter/scc/v3@latest
    github.com/Antonboom/testifylint@latest
)

# Packages needing more memory to build than some Raspberry Pi can provide
MINIMUM_RAM_MB=1500
FAT_PACKAGES=(
    github.com/nats-io/natscli/nats@latest
    github.com/sqlc-dev/sqlc/cmd/sqlc@latest
)

# Install each of the given packages in turn.
install_packages() {
    local package
    for package in "$@"; do
        echo "==> go install $package"
        go install "$package"
    done
}

# Install core set of packages
install_packages "${PACKAGES[@]}"

# Install 'fat' packages?
ram_mb=$(( $(awk '/^MemTotal:/ { print $2 }' /proc/meminfo) / 1024 ))
if (( ram_mb >= MINIMUM_RAM_MB )); then
    install_packages "${FAT_PACKAGES[@]}"
else
    echo
    echo "Only ${ram_mb}MB RAM, need ${MINIMUM_RAM_MB}MB: skipping ${#FAT_PACKAGES[@]} large packages"
fi

echo
echo "Clear all Go caches"
go clean -cache -testcache -modcache -fuzzcache
