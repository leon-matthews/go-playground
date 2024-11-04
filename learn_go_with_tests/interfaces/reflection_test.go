package main

import (
	"slices"
	"testing"
)

type Person struct {
	Name    string
	Profile Profile
}

type Profile struct {
	City     string
	Postcode uint
}

func TestWalk(t *testing.T) {
	cases := []struct {
		Name          string
		Input         interface{}
		ExpectedCalls []string
	}{
		{
			"struct with one string field",
			struct {
				Name string
			}{"Leon"}, []string{"Leon"},
		},
		{
			"struct with two string fields",
			struct {
				Name, City string
			}{"Leon", "Auckland"},
			[]string{"Leon", "Auckland"},
		},
		{
			"struct with an integer field",
			struct {
				Name     string
				Postcode int
			}{"Leon", 1026},
			[]string{"Leon"},
		},
		{
			"nested fields",
			Person{
				"Leon",
				Profile{"Auckland", 1026},
			},
			[]string{"Leon", "Auckland"},
		},
		{
			"pointers to things",
			&Person{
				"Leon",
				Profile{"Auckland", 1206},
			},
			[]string{"Leon", "Auckland"},
		},
	}

	for _, test := range cases {
		t.Run(test.Name, func(t *testing.T) {
			var got []string
			walk(test.Input, func(input string) {
				got = append(got, input)
			})

			if !slices.Equal(got, test.ExpectedCalls) {
				t.Errorf("got %v, want %v", got, test.ExpectedCalls)
			}
		})
	}
}
