package main

import (
	"testing"

	"github.com/alecthomas/assert/v2"
	"pgregory.net/rapid"
)

var tests = []struct {
	Name   string
	Arabic int
	Roman  string
}{
	{"one", 1, "I"},
	{"two", 2, "II"},
	{"four", 4, "IV"},
	{"five", 5, "V"},
	{"six", 6, "VI"},
	{"seven", 7, "VII"},
	{"eight", 8, "VIII"},
	{"nine", 9, "IX"},
	{"ten", 10, "X"},
	{"eleven", 11, "XI"},
	{"fourteen", 14, "XIV"},
	{"eightteen", 18, "XVIII"},
	{"twenty", 20, "XX"},
	{"thirty-nine", 39, "XXXIX"},
	{"forty", 40, "XL"},
	{"forty-seven", 47, "XLVII"},
	{"forty-nine", 49, "XLIX"},
	{"fifty", 50, "L"},
	{"ninety-nine", 99, "XCIX"},
	{"one-hundred", 100, "C"},
	{"four-hundred and ninety-nine", 499, "CDXCIX"},
	{"three-thousand nine hundred and ninety-nine", 3999, "MMMCMXCIX"},
}

func TestConvertToRoman(t *testing.T) {
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			got := ConvertToRoman(tt.Arabic)
			assert.Equal(t, tt.Roman, got)
		})
	}
}

func TestConvertToArabic(t *testing.T) {
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			got := ConvertToArabic(tt.Roman)
			assert.Equal(t, tt.Arabic, got)
		})
	}
}

func conversionRoundTrip(t *rapid.T) {
	arabic := rapid.IntRange(1, 3999).Draw(t, "arabic")
	roman := ConvertToRoman(arabic)
	got := ConvertToArabic(roman)
	if got != arabic {
		t.Fatalf("ConvertToArabic(%q) = %d, want %d", roman, got, arabic)
	}
}

func TestConversionProperties(t *testing.T) {
	rapid.Check(t, conversionRoundTrip)
}

func FuzzRoundTrip(f *testing.F) {
	f.Fuzz(rapid.MakeFuzz(conversionRoundTrip))
}
