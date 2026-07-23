package humanise_test

import (
	"fmt"
	"time"

	"local.dev/humanise"
)

func Example() {
	fmt.Println(humanise.Comma(1234567))
	fmt.Println(humanise.Words(1200000))
	fmt.Println(humanise.Ordinal(21))
	fmt.Println(humanise.FileSize(4200))
	// Output:
	// 1,234,567
	// 1.2 million
	// 21st
	// 4.2kB
}

func ExampleComma() {
	fmt.Println(humanise.Comma(1234567))
	fmt.Println(humanise.Comma(-42))
	// Output:
	// 1,234,567
	// -42
}

func ExampleUnderscore() {
	fmt.Println(humanise.Underscore(1234567))
	// Output: 1_234_567
}

func ExampleWords() {
	fmt.Println(humanise.Words(1200000))
	fmt.Println(humanise.Words(16000))
	fmt.Println(humanise.Words(1999)) // below 2000 stays a comma literal
	// Output:
	// 1.2 million
	// 16 thousand
	// 1,999
}

func ExampleWordsCompact() {
	fmt.Println(humanise.WordsCompact(1200000))
	fmt.Println(humanise.WordsCompact(1500))
	fmt.Println(humanise.WordsCompact(999)) // below 1000 stays plain digits
	// Output:
	// 1.2M
	// 1.5K
	// 999
}

func ExampleOrdinal() {
	for _, n := range []int64{1, 2, 3, 11, 21, 113} {
		fmt.Printf("%s ", humanise.Ordinal(n))
	}
	fmt.Println()
	// Output: 1st 2nd 3rd 11th 21st 113th
}

func ExampleFileSize() {
	for _, size := range []int64{512, 4200, 1_500_000} {
		fmt.Println(humanise.FileSize(size))
	}
	// Output:
	// 512B
	// 4.2kB
	// 1.5MB
}

func ExampleFileSizeIEC() {
	fmt.Println(humanise.FileSizeIEC(4200))
	// Output: 4.1KiB
}

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

func ExampleDuration() {
	fmt.Println(humanise.Duration(3 * time.Hour))
	fmt.Println(humanise.Duration(90 * time.Second)) // one minute drops to seconds
	// Output:
	// 3 hours
	// 90 seconds
}

func ExampleRelative() {
	fmt.Println(humanise.Relative(-5 * time.Minute))
	fmt.Println(humanise.Relative(3 * 24 * time.Hour))
	fmt.Println(humanise.Relative(0))
	// Output:
	// 5 minutes ago
	// in 3 days
	// now
}

func ExampleAge() {
	born := time.Date(1976, time.February, 1, 0, 0, 0, 0, time.UTC)
	today := time.Date(2022, time.July, 4, 0, 0, 0, 0, time.UTC)
	fmt.Println(humanise.Age(born, today))
	// Output: 46
}

func ExampleSignificant() {
	fmt.Println(humanise.Significant(1234.567, 3))
	fmt.Println(humanise.Significant(0.0001234, 2))
	// Output:
	// 1230
	// 0.00012
}

func ExampleAnd() {
	fmt.Println(humanise.And([]string{"apples", "oranges", "bananas"}))
	fmt.Println(humanise.And([]string{"apples", "oranges"})) // two items take no comma
	// Output:
	// apples, oranges, and bananas
	// apples and oranges
}

func ExampleOr() {
	fmt.Println(humanise.Or([]string{"apples", "oranges", "bananas"}))
	// Output: apples, oranges, or bananas
}

func ExampleTitle() {
	fmt.Println(humanise.Title("taming of the shrew"))
	fmt.Println(humanise.Title("Taming Of The Shrew")) // capitalised minor words are tidied
	// Output:
	// Taming of the Shrew
	// Taming of the Shrew
}
