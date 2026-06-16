// Findlinks does an HTTP GET on each URL, parses the
// result as HTML, and prints the links within it.
//
// Usage:
//
//	findlinks url ...
//
// Taken from: https://github.com/adonovan/gopl.io/tree/master/ch5/findlinks2
package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/tdewolff/parse/v2"
	tdehtml "github.com/tdewolff/parse/v2/html"
	"golang.org/x/net/html"
)

func main() {
	urls := os.Args[1:]
	if len(urls) == 0 {
		fmt.Fprintln(os.Stderr, "usage: findlinks URL...")
		os.Exit(1)
	}

	for _, url := range urls {
		// Download HTML
		body, err := fetchBody(url)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Fetched %d bytes from %s\n", len(body), url)

		// Time the tokenizer, capturing elapsed before logging so I/O stays out of it
		start := time.Now()
		links, err := tokenizeLinks(body)
		elapsed := time.Since(start)
		if err != nil {
			log.Fatal(fmt.Errorf("findLinks: %w", err))
		}
		log.Printf("Found %d links by tokenising in %v\n", len(links), elapsed)

		// Time the parser the same way
		start = time.Now()
		links, numCalls, err := parseLinks(body)
		elapsed = time.Since(start)
		if err != nil {
			log.Fatal(fmt.Errorf("findLinks: %w", err))
		}
		log.Printf("Found %d links by parsing in %v (%d parseVisit calls)\n", len(links), elapsed, numCalls)

		printLinks(links, "")
	}
}

func fetchBody(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("getting %s: %s", url, resp.Status)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading %s: %s", url, err)
	}
	return string(b), nil
}

// tokenizeLinks() extracts the links using just tokens
func tokenizeLinks(body string) ([]string, error) {
	var links []string
	z := html.NewTokenizer(strings.NewReader(body))
	for {
		switch z.Next() {
		case html.ErrorToken:
			// ErrorToken ends the stream; io.EOF is the clean finish, anything else is real
			if z.Err() == io.EOF {
				return links, nil
			}
			return links, z.Err()
		case html.StartTagToken, html.SelfClosingTagToken:
			name, hasAttr := z.TagName()
			if hasAttr && string(name) == "a" {
				for {
					// key/val are []byte into the buffer
					key, val, more := z.TagAttr()
					if string(key) == "href" {
						href := string(val)
						links = append(links, href)
					}
					if !more {
						break
					}
				}
			}
		}
	}
}

// tdewolffLinks extracts the links using the tdewolff/parse HTML lexer.
func tdewolffLinks(body string) ([]string, error) {
	var links []string
	lexer := tdehtml.NewLexer(parse.NewInputString(body))
	inAnchor := false
	for {
		tt, _ := lexer.Next()
		switch tt {
		case tdehtml.ErrorToken:
			// ErrorToken ends the stream; io.EOF is the clean finish, anything else is real
			if lexer.Err() == io.EOF {
				return links, nil
			}
			return links, lexer.Err()
		case tdehtml.StartTagToken:
			// Attributes that follow belong to this tag, so remember if it is an anchor
			inAnchor = string(lexer.Text()) == "a"
		case tdehtml.AttributeToken:
			if inAnchor && string(lexer.AttrKey()) == "href" {
				links = append(links, attrValue(lexer.AttrVal()))
			}
		}
	}
}

// attrValue strips surrounding quotes and decodes HTML entities from a raw lexer value.
func attrValue(raw []byte) string {
	if n := len(raw); n >= 2 && (raw[0] == '"' || raw[0] == '\'') && raw[n-1] == raw[0] {
		raw = raw[1 : n-1]
	}
	return html.UnescapeString(string(raw))
}

// indexLinks does a rough strings.Index scan for `<a ` tags; double-quoted href only, no unescaping.
func indexLinks(body string) ([]string, error) {
	var links []string
	for {
		i := strings.Index(body, "<a ")
		if i < 0 {
			break
		}
		body = body[i+len("<a "):]

		// Isolate the tag so href searches can't run past the closing '>'
		end := strings.IndexByte(body, '>')
		if end < 0 {
			break
		}
		tag := body[:end]
		body = body[end:]

		h := strings.Index(tag, `href="`)
		if h < 0 {
			continue
		}
		value := tag[h+len(`href="`):]
		if q := strings.IndexByte(value, '"'); q >= 0 {
			links = append(links, value[:q])
		}
	}
	return links, nil
}

// anchorHref matches an <a> tag's double-quoted href value.
var anchorHref = regexp.MustCompile(`<a [^>]*href="([^"]*)"`)

// regexpLinks extracts <a> tag links with a best-effort regexp, decoding entities.
func regexpLinks(body string) ([]string, error) {
	matches := anchorHref.FindAllStringSubmatch(body, -1)
	links := make([]string, 0, len(matches))
	for _, m := range matches {
		links = append(links, html.UnescapeString(m[1]))
	}
	return links, nil
}

// parseLinks extracts the links by parsing the entire HTML document.
//
// It also reports how many parseVisit calls the recursive walk took.
func parseLinks(body string) (links []string, numCalls int, err error) {
	node, err := html.Parse(strings.NewReader(body))
	if err != nil {
		return nil, 0, fmt.Errorf("parsing HTML: %w", err)
	}
	links = parseVisit(nil, node, &numCalls)
	return links, numCalls, nil
}

// parseVisit recursively appends to links each link found in n
func parseVisit(links []string, n *html.Node, numCalls *int) []string {
	*numCalls++

	// Try and find href attr in <a> element
	if n.Type == html.ElementNode && n.Data == "a" {
		for _, a := range n.Attr {
			if a.Key == "href" {
				links = append(links, a.Val)
			}
		}
	}

	// Recursively visit all child nodes
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		links = parseVisit(links, c, numCalls)
	}
	return links
}

// printLinks deduplicates, sorts, and prints links with given indent
func printLinks(links []string, indent string) {
	slices.Sort(links)
	links = slices.Compact(links)
	for _, url := range links {
		fmt.Printf("%s%s\n", indent, url)
	}
}
