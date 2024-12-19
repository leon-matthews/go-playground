
# My Go Setup

This is my ongoing collection of shell aliases and editor snippets for the
Go programming language.


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

    export GOPATH=$HOME/.go
    FOLDER="$GOPATH/bin"
    if [ -d "$FOLDER" ] ; then
        PATH="$FOLDER:$PATH"
    fi

## Go tools

### errcheck

The `errcheck` linter checks for un-inspected error return values:

    $ go install github.com/kisielk/errcheck@latest


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
