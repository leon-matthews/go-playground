package main

import (
	"os"
	"text/template"
)

func main() {
	// Templates are a mix of static text and “actions” enclosed in {{...}}
	t1 := template.New("t1")
	t1, err := t1.Parse("Value is {{.}}\n")
	if err != nil {
		panic(err)
	}

	// Like above, creating using 'must' will panic if parse returns error
	t1 = template.Must(t1.Parse("Value: {{.}}\n"))

	// Executing the template generates its text with specific values
	t1.Execute(os.Stdout, "some text")
	t1.Execute(os.Stdout, 5)
	t1.Execute(os.Stdout, []string{"apple", "banana"})

	// If the data is a struct we can use the {{.FieldName}} action on exported fields
	t2 := Create("t2", "Name: {{.Name}}\n")
	t2.Execute(os.Stdout, struct {
		Name string
	}{"Jane Doe"})

	// The same applies to maps, note that keys are case-sensitive
	t2.Execute(os.Stdout, map[string]string{
		"Name": "John Q Public",
	})

	// if/else provides conditional execution. Zero values are considered false.
	t3 := Create("t3", "{{if . -}} yes {{else -}} no {{end}}\n")
	t3.Execute(os.Stdout, "not empty")
	t3.Execute(os.Stdout, "")

	// range blocks will loop through slices, arrays, maps or channels
	t4 := Create("t4", "Range: {{range .}}{{.}} {{end}}\n")
	t4.Execute(os.Stdout, []string{"Go", "Rust", "C++", "Python"})
}

// Create makes and parses template, panicking on error
func Create(name, t string) *template.Template {
	return template.Must(template.New(name).Parse(t))
}
