package main

import (
	"slices"
	"testing"
)

type Person struct {
	Name string
	Profile
}

type Profile struct {
	Age  int
	City string
}

func TestWalk(t *testing.T) {
	tests := map[string]struct {
		input any
		want  []string
	}{
		"struct with one string field": {
			struct{ Status string }{"Single"},
			[]string{"Single"},
		},
		"struct with two string fields": {
			struct{ First, Last string }{"John", "Smith"},
			[]string{"John", "Smith"},
		},
		"struct with a non-string field": {
			Profile{50, "Auckland"},
			[]string{"Auckland"},
		},
		"pointer to struct": {
			&Profile{68, "Tauranga"},
			[]string{"Tauranga"},
		},
		"nested fields": {
			Person{
				"Leon",
				Profile{
					50,
					"Auckland",
				},
			},
			[]string{"Leon", "Auckland"},
		},
		"slices": {
			[]Profile{
				{50, "Auckland"},
				{68, "Tauranga"},
			},
			[]string{"Auckland", "Tauranga"},
		},
		"arrays": {
			[2]Profile{
				{50, "Auckland"},
				{68, "Tauranga"},
			},
			[]string{"Auckland", "Tauranga"},
		},
		"maps": {
			map[string]string{
				"cow":   "Moo!",
				"sheep": "Baa!",
			},
			[]string{"Moo!", "Baa!"},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			var got []string
			walk(tt.input, func(input string) {
				got = append(got, input)
			})

			if !slices.Equal(got, tt.want) {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}
