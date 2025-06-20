package blogposts

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"
)

const (
	titleSeparator       = "Title: "
	descriptionSeparator = "Description: "
	tagSeparator         = "Tags: "
)

func parsePost(postFile io.Reader) (Post, error) {
	scanner := bufio.NewScanner(postFile)

	readMetaline := func(separator string) string {
		scanner.Scan()
		return strings.TrimPrefix(scanner.Text(), separator)
	}

	title := readMetaline(titleSeparator)
	description := readMetaline(descriptionSeparator)
	tags := strings.Split(readMetaline(tagSeparator), ", ")
	body := readBody(scanner)

	post := Post{Title: title, Description: description, Tags: tags, Markdown: body}
	return post, nil
}

func readBody(scanner *bufio.Scanner) string {
	buf := bytes.Buffer{}
	// Skip separator
	scanner.Scan()

	// Read body, line by line
	for scanner.Scan() {
		fmt.Fprintln(&buf, scanner.Text())
	}
	body := strings.TrimSuffix(buf.String(), "\n")
	return body
}
