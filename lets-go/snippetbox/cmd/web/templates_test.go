package main

import (
	"testing"
	"time"

	"local.dev/snippetbox/internal/assert"
)

func TestHumanDate(t *testing.T) {
	nz, err := time.LoadLocation("Pacific/Auckland")
	if err != nil {
		t.Fatal("Could not load timezone")
	}

	tests := map[string]struct {
		then time.Time
		want string
	}{
		"UTC": {
			then: time.Date(2024, 3, 17, 10, 15, 0, 0, time.UTC),
			want: "17 Mar 2024 at 10:15",
		},
		"Empty": {
			then: time.Time{},
			want: "",
		},
		"NZ": {
			then: time.Date(2024, 3, 17, 10, 15, 0, 0, nz),
			want: "16 Mar 2024 at 21:15",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := humanDate(tt.then)
			assert.Equal(t, got, tt.want)
		})
	}
}
