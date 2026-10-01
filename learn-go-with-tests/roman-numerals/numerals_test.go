package main

import (
	"testing"
	"testing/quick"

	"github.com/alecthomas/assert/v2"
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

func TestConversionProperties(t *testing.T) {
	roundTrip := func(arabic int) bool {
		roman := ConvertToRoman(int(arabic))
		fromRoman := ConvertToArabic(roman)
		return fromRoman == arabic
	}

	if err := quick.Check(roundTrip, nil); err != nil {
		t.Error("failed checks:", err)
	}
}
