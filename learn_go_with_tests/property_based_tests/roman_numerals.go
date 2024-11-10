package main

import (
	"fmt"
	"strings"
)

func main() {
	fmt.Println("property_based_tests")
}

type RomanNumeral struct {
	Value  uint
	Symbol string
}

var allRomanNumerals = []RomanNumeral{
	{1_000_000, "M̅"},
	{9_00_000, "C̅M̅"},
	{500_000, "D̅"},
	{400_000, "C̅D̅"},
	{100_000, "C̅"},
	{90_000, "X̅C̅"},
	{50_000, "L̅"},
	{40_000, "X̅L̅"},
	{10_000, "X̅"},
	{9_000, "I̅X̅"},
	{5_000, "V̅"},
	{4_000, "I̅"},
	{1_000, "M"},
	{900, "CM"},
	{500, "D"},
	{400, "CD"},
	{100, "C"},
	{90, "XC"},
	{50, "L"},
	{40, "XL"},
	{10, "X"},
	{9, "IX"},
	{5, "V"},
	{4, "IV"},
	{1, "I"},
}

func ConvertToRoman(arabic uint) string {
	var result strings.Builder

	for _, numeral := range allRomanNumerals {
		for arabic >= numeral.Value {
			result.WriteString(numeral.Symbol)
			arabic -= numeral.Value
		}
	}

	return result.String()
}

func ConvertToArabic(roman string) uint {
	var arabic uint = 0

	for _, numeral := range allRomanNumerals {
		for strings.HasPrefix(roman, numeral.Symbol) {
			arabic += numeral.Value
			roman = strings.TrimPrefix(roman, numeral.Symbol)
		}
	}

	return arabic
}
