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
			Name: "struct with one string field",
			Input: struct {
				Name string
			}{"Leon"}, ExpectedCalls: []string{"Leon"},
		},
		{
			Name: "struct with two string fields",
			Input: struct {
				Name, City string
			}{"Leon", "Auckland"},
			ExpectedCalls: []string{"Leon", "Auckland"},
		},
		{
			Name: "struct with an integer field",
			Input: struct {
				Name     string
				Postcode int
			}{"Leon", 1026},
			ExpectedCalls: []string{"Leon"},
		},
		{
			Name: "nested fields",
			Input: Person{
				"Leon",
				Profile{"Auckland", 1026},
			},
			ExpectedCalls: []string{"Leon", "Auckland"},
		},
		{
			Name: "pointers to things",
			Input: &Person{
				"Leon",
				Profile{"Auckland", 1206},
			},
			ExpectedCalls: []string{"Leon", "Auckland"},
		},
		{
			Name: "slices",
			Input: []Profile{
				{"Waterview", 1026},
				{"Avondale", 1032},
			},
			ExpectedCalls: []string{"Waterview", "Avondale"},
		},
		{
			Name: "arrays",
			Input: [2]Profile{
				{"Waterview", 1026},
				{"Avondale", 1032},
			},
			ExpectedCalls: []string{"Waterview", "Avondale"},
		},
		{
			"maps",
			map[string]string{
				"Cow":   "Moo",
				"Sheep": "Baa",
			},
			[]string{"Moo", "Baa"},
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
