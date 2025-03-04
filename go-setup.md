
# My Go Setup

This is my ongoing collection of shell aliases and editor snippets for the
Go programming language.

## Install manually


    $ rm -rf /usr/local/go
    $ tar -C /usr/local -xzf go1.24.0.linux-amd64.tar.gz


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

    # Go!
    export GOPATH=$HOME/.go
    FOLDER="$GOPATH/bin"
    if [ -d "$FOLDER" ] ; then
        PATH="$FOLDER:$PATH"
    fi

## Go tools


### cobra-cli

Cobra Generator generates the scaffolding for new CLI application.

    $ go install github.com/spf13/cobra-cli@latest


### delve

Delve debugger works better than GDB for Go programs.

    $ go install github.com/go-delve/delve/cmd/dlv@latest


### errcheck

The `errcheck` linter checks for un-inspected error return values:

    $ go install github.com/kisielk/errcheck@latest


### go-size-analyzer

Explore which dependencies are making your binary large.

    $ go install github.com/Zxilly/go-size-analyzer/cmd/gsa@latest
    $ gsa hello-world


### gotestsum

Reformat test output, and automatically run tests after code changes
with `gotestsum`:

    $ go install gotest.tools/gotestsum@latest
    $ gotestsum
    $ gotestsum
    ✓  . (cached)
    DONE 4 tests in 0.050s


    $ gotestsum --watch

    $ gotestsum -f dots-v2
    $ gotestsum -f dots-v2 -- -run JustOneTest


### revive

Go linter

    $ go install github.com/mgechev/revive@latest

### staticcheck

Another Go linter

    $ go install honnef.co/go/tools/cmd/staticcheck@latest
    $ staticcheck ./...


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
