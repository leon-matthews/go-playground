package humanise_test

import (
	"fmt"
	"time"

	"local.dev/humanise"
)

// Example gives an at-a-glance tour of the package.
func Example() {
	fmt.Println(humanise.Comma(1234567))
	fmt.Println(humanise.Words(1200000))
	fmt.Println(humanise.Ordinal(21))

	size, _ := humanise.FileSize(4200)
	fmt.Println(size)
	// Output:
	// 1,234,567
	// 1.2 million
	// 21st
	// 4.2kB
}

// ExampleComma groups an integer with thousands separators for display.
func ExampleComma() {
	fmt.Println(humanise.Comma(1234567))
	fmt.Println(humanise.Comma(-42))
	// Output:
	// 1,234,567
	// -42
}

// ExampleUnderscore groups an integer the way a Go numeric literal is written.
func ExampleUnderscore() {
	fmt.Println(humanise.Underscore(1234567))
	// Output: 1_234_567
}

// ExampleWords renders large integers with a long-scale word.
func ExampleWords() {
	fmt.Println(humanise.Words(1200000))
	fmt.Println(humanise.Words(16000))
	fmt.Println(humanise.Words(1999)) // below 2000 stays a comma literal
	// Output:
	// 1.2 million
	// 16 thousand
	// 1,999
}

// ExampleWordsCompact renders large integers with a short-scale suffix.
func ExampleWordsCompact() {
	fmt.Println(humanise.WordsCompact(1200000))
	fmt.Println(humanise.WordsCompact(1500))
	fmt.Println(humanise.WordsCompact(999)) // below 1000 stays plain digits
	// Output:
	// 1.2M
	// 1.5K
	// 999
}

// ExampleOrdinal renders integers as English ordinals, including the 11-13 teens.
func ExampleOrdinal() {
	for _, n := range []int64{1, 2, 3, 11, 21, 113} {
		fmt.Printf("%s ", humanise.Ordinal(n))
	}
	fmt.Println()
	// Output: 1st 2nd 3rd 11th 21st 113th
}

// ExampleFileSize renders byte counts with SI (base-1000) units.
func ExampleFileSize() {
	for _, size := range []float64{512, 4200, 1_500_000} {
		formatted, _ := humanise.FileSize(size)
		fmt.Println(formatted)
	}
	// Output:
	// 512B
	// 4.2kB
	// 1.5MB
}

// ExampleFileSizeIEC renders byte counts with IEC (base-1024) units.
func ExampleFileSizeIEC() {
	formatted, _ := humanise.FileSizeIEC(4200)
	fmt.Println(formatted)
	// Output: 4.1KiB
}

// ExampleMetric renders values with an SI prefix and unit.
func ExampleMetric() {
	readings := []struct {
		value float64
		unit  string
	}{
		{1500, "V"},
		{0.005, "A"},
		{2e8, "W"},
	}
	for _, r := range readings {
		formatted, _ := humanise.Metric(r.value, r.unit)
		fmt.Println(formatted)
	}
	// Output:
	// 1.5 kV
	// 5 mA
	// 200 MW
}

// ExampleDuration renders a time span as an approximate phrase.
func ExampleDuration() {
	fmt.Println(humanise.Duration(3 * time.Hour))
	fmt.Println(humanise.Duration(90 * time.Second)) // one minute drops to seconds
	// Output:
	// 3 hours
	// 90 seconds
}

// ExampleRelative renders a signed offset as a past or future phrase.
func ExampleRelative() {
	fmt.Println(humanise.Relative(-5 * time.Minute))
	fmt.Println(humanise.Relative(3 * 24 * time.Hour))
	fmt.Println(humanise.Relative(0))
	// Output:
	// 5 minutes ago
	// in 3 days
	// now
}

// ExampleSignificant rounds a value to a number of significant figures.
func ExampleSignificant() {
	fmt.Println(humanise.Significant(1234.567, 3))
	fmt.Println(humanise.Significant(0.0001234, 2))
	// Output:
	// 1230
	// 0.00012
}
