// Books sorts and prints a collection of books.
package main

import (
	"cmp"
	"fmt"
	"os"
	"slices"
	"strings"
	"text/tabwriter"
)

type Authors []string

func (authors Authors) String() string {
	return strings.Join(authors, ", ")
}

type Book struct {
	Authors
	Title string
	Year  int
}

func printBooks(books []Book) {
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer tw.Flush()
	format := "%v\t%v\t%v\n"
	fmt.Fprintf(tw, format, "Authors", "Title", "Year")
	fmt.Fprintf(tw, format, "-------", "-----", "----")
	for _, b := range books {
		fmt.Fprintf(tw, format, b.Authors, b.Title, b.Year)
	}
}

func main() {
	books := []Book{
		{Authors{"Tolkien"}, "The Lord of the Rings", 1954},
		{Authors{"Kernighan", "Donovan"}, "The Go Programming Language", 2015},
		{Authors{"Kim", "Behr", "Spafford"}, "The Phoenix Project", 2013},
	}
	slices.SortFunc(books, func(a, b Book) int {
		return cmp.Compare(a.Year, b.Year)
	})
	printBooks(books)
}
