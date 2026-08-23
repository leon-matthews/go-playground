
# My Go Setup

This is my ongoing collection of shell aliases and editor snippets for the
Go programming language.

## Install manually


	$ wget
    $ sudo rm -rf /usr/local/go
    $ sudo tar -C /usr/local -xzf go1.24.4.linux-amd64.tar.gz


## Bash aliases

Shell aliases for command-line quickness! Add to your *.bashrc* or equivalent.

    alias gob="go test -bench ."
    alias gof="go fmt ./..."
    alias goi="goimports -l -w ."
    alias got="go test"
    alias gov="go vet ./..."
    alias goc="go clean -r -cache -fuzzcache -modcache -testcache"


## Bash setup

Install Go stuff into a dotfolder and add its *bin* directory to `$PATH`.

In your *.profile* or similar:

# GOPATH
export GOPATH="$HOME/.go"
FOLDER="$GOPATH/bin"
if [ -d "$FOLDER" ] ; then
    PATH="$FOLDER:$PATH"
fi

# Go installation
if [ -d /usr/local/go/bin ] ; then
    PATH="$PATH:/usr/local/go/bin"
fi


## Go tools

### bbolt

CLI for bbolt embedded key/value database

    $ go install go.etcd.io/bbolt/cmd/bbolt@latest

### benchstat

Benchstat computes statistical summaries and A/B comparisons of Go benchmarks. 

    $ go install golang.org/x/perf/cmd/benchstat@latest

### glow

Browse and read markdown documentation under current folder:

	$ go install github.com/charmbracelet/glow@latest

### gofupmt

A stricter gofmt

    $ go install mvdan.cc/gofumpt@latest

### goimports

Command goimports updates your Go import lines, adding missing ones and removing unreferenced ones.

	$ go install golang.org/x/tools/cmd/goimports@latest

### golangci-lint

Runs various linters in parallel, collecting their results.
To ensure that all denpendencies are at the version they've testing do not use the `latest` tag.

    $ go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.13.1

### go-mod-upgrade

Check for updates to 3rd-party packages automatically.

    $ go install github.com/oligot/go-mod-upgrade@latest

### gopls

The official language server for Go, provides IDE features to any 
LSP-compatible editor.

    $ go install golang.org/x/tools/gopls@latest

### govulncheck

Find security issues by scanning your project's dependencies for known 
vulnerabilities and then identifying any direct or indirect calls to same.

    $ go install golang.org/x/vuln/cmd/govulncheck@latest

### go-size-analyzer

Explore which dependencies are making your binary large.

    $ go install github.com/Zxilly/go-size-analyzer/cmd/gsa@latest
    $ gsa hello-world


### gotestsum

Reformat test output, and automatically run tests after code changes
with `gotestsum`:

    $ go install gotest.tools/gotestsum@latest
    $ gotestsum
    ✓  . (cached)
    DONE 4 tests in 0.050s


    $ gotestsum --watch

    $ gotestsum -f dots-v2
    $ gotestsum -f dots-v2 -- -run JustOneTest

### NATS

CLI to interact with NATS

    $ go install github.com/nats-io/natscli/nats@latest

### NilAway

Satic analysis tool that attempts to find potential nil panics.

	$ go install go.uber.org/nilaway/cmd/nilaway@latest

### revive

Very fast, stand-alone Go linter with good defaults.

    $ go install github.com/mgechev/revive@latest
    $ revive ./...

### scc

Sloc, Cloc and Code: scc is a fast code counter.

    $ go install github.com/boyter/scc/v3@latest

### sqlc

Generates fully type-safe idiomatic Go code from SQL.

    $ go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

### testifylint

Improve usage of testing library `github.com/stretchr/testify`

    $ go install github.com/Antonboom/testifylint@latest
    $ testifylint ./...


## Editor snippets

Like shell aliases, I like to automate my text editor to emit common
boilerplate code. In the following section `%cursor%` is where your cursor
ends up after hitting *<tab>*.

    # General
    d=fmt.Printf("[%T]%+[1]v\\n", %cursor%)
    err=if err != nil {\n\t%cursor%\n}
    p=fmt.Println(%cursor%)

    # Testing
    btest=func Benchmark%cursor%(b *testing.B) {\n\tfor i := 0; i < b.N; i++ {}\n}
    htest=func assert%cursor%(t testing.TB) {\n\tt.Helper()\n}
    stest=t.Run("%cursor%", func(t *testing.T) {\n})
    test=func Test%cursor%(t *testing.T) {\n}
