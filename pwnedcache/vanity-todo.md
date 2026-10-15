# Vanity import path: lost.co.nz/pwnedcache

Goal: `go install lost.co.nz/pwnedcache@latest` installs from the Codeberg repository.


## 1. Prepare the project

- Create the repository at `https://codeberg.org/leon_matthews/pwnedcache`.
- Change the module path from `local.dev/pwnedcache` to `lost.co.nz/pwnedcache` in `go.mod` and
  in every internal import.
- Fill in the `go install` line in `README.md`, and the project URL in the User-Agent.


## 2. Serve the meta tag from lost.co.nz

The `go` command sends `GET https://lost.co.nz/pwnedcache?go-get=1` and reads the `<head>` of
the HTML response for:

    <meta name="go-import" content="lost.co.nz/pwnedcache git https://codeberg.org/leon_matthews/pwnedcache">

- The first field must exactly match the `module` line in `go.mod`.
- The repository URL needs a scheme and no `.git` suffix.
- Put the tag early in `<head>`, before any CSS or JavaScript: the `go` command's HTML parser
  is restricted.
- The response status is ignored and redirects are followed; only the meta tag matters.
- Answer `/pwnedcache` and any `/pwnedcache/...` subpath with the same tag. For a subpath, the
  `go` command makes a second request to the prefix to check that the tags match.
- HTTPS only. Plain HTTP is used only by clients that set `GOINSECURE`.
- Requests usually come from `proxy.golang.org`, not the user's machine, so the page must be
  public: no login, cookie wall, or bot challenge.
- Requests without `?go-get=1` can be served anything, such as a project page or a redirect to
  the repository.

Django: match `pwnedcache` without a trailing slash, so `APPEND_SLASH` never redirects it.


## 3. Verify, before tagging anything

    $ curl -s 'https://lost.co.nz/pwnedcache?go-get=1' | grep go-import
    $ GOPROXY=direct go install lost.co.nz/pwnedcache@latest

`GOPROXY=direct` makes the request from your own machine instead of the proxy. Until the
repository has a tagged commit, `@latest` resolves to a pseudo-version.


## 4. Release

    $ git tag v0.1.0
    $ git push origin v0.1.0
    $ go install lost.co.nz/pwnedcache@v0.1.0

The last command asks the proxy for the new version; it should appear within a minute, but can
take up to 30 minutes if that version was requested before the tag was pushed. pkg.go.dev picks
new versions up from the proxy within a few minutes.

- **Never move or delete a pushed tag.** The proxy and checksum database keep the first copy
  they see, and a changed tag gives users a checksum security error.
- To withdraw a bad release, add a `retract` directive to `go.mod` and publish a newer version.


## 5. Optional: source links on pkg.go.dev

If pkg.go.dev's source links for the package don't work, add a `go-source` meta tag alongside
`go-import` (format described on pkg.go.dev/about).


## Sources

- `go help importpath`
- Go Modules Reference: <https://go.dev/ref/mod>, sections "Finding a repository for a module
  path", "Privacy", "Mapping versions to commits", and "`retract` directive"
- Go module proxy FAQ: <https://proxy.golang.org/>
- pkg.go.dev: <https://pkg.go.dev/about>
- Django `APPEND_SLASH`: <https://docs.djangoproject.com/en/stable/ref/settings/#append-slash>
