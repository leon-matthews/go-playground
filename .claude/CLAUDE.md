# Go Playground Context & Rules

This repo is a collection of Go learning resources and Go projects that are
just in the exploration stage. Code that has gestated for a while is expected
to be production-ready.


## Commands to validate Go code

If a folder contains a `go.mod` file then it is a module that contains one or
more proper Go packages. If it's a bare Go file, eg. `main.go` it's just a
single-file script designed to be run with eg. `go run main.go`.

In either case, don't run `go build` to validate syntax, as this leaves large
binaries behind and is not necessary.

### Validate packages

Run these from inside the module folder (the one containing `go.mod`); the repo root is not
a module, so `./...` fails there.

- Check syntax, basic errors: `go vet ./...`
- Check for use of modern idioms: `go fix --diff ./...`
- Basic linting: `revive ./...`
- Apply automatic formatting (may modify source files): `gofumpt --extra -w .`

### Validate single files

Assuming that the single file is called `hello.go`

- Check syntax, basic errors: `go vet hello.go`
- Check for use of modern idioms: `go fix --diff hello.go`
- Basic linting: `revive hello.go`
- Apply automatic formatting (may modify source files): `gofumpt --extra -w hello.go`


## Claude Code CLI Constraints

- Keep all shared memory, tools, and rules locked to this root file.
- Use a plain hyphen character, not an mdash.
