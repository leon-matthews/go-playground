# Snippetbox

Code sharing website demo from Alex Edwards' book "Let's Go".

## Development Server

Use `go run` to run HTTP server locally:

```bash
$ go run ./cmd/web/ --addr=localhost:8000
```

To automatically restart server after code changes I like to use the CLI
utility [Reflex](https://github.com/cespare/reflex), as so:


```bash
$ reflex -s -- sh -c 'go run ./cmd/web/ --addr=localhost:8000'
```
