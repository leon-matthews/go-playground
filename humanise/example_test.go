package humanise_test

import (
	"fmt"
	"math"
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
	// 4.2 kB
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
	fmt.Println(humanise.Words(1_000_000_000))
	fmt.Println(humanise.Words(16000))
	fmt.Println(humanise.Words(1999)) // below 2000 stays a comma literal
	// Output:
	// 1.2 million
	// 1 billion
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
	// The last value shows that negatives render by magnitude.
	for _, n := range []int64{1, 2, 3, 11, 21, 113, -21} {
		fmt.Printf("%s ", humanise.Ordinal(n))
	}
	fmt.Println()
	// Output: 1st 2nd 3rd 11th 21st 113th 21st
}

func ExampleFileSize() {
	for _, size := range []int64{512, 4200, 1_500_000} {
		fmt.Println(humanise.FileSize(size))
	}
	// Output:
	// 512 B
	// 4.2 kB
	// 1.5 MB
}

func ExampleFileSizeIEC() {
	// The same sizes as ExampleFileSize, in binary multiples.
	for _, size := range []int64{512, 4200, 1_500_000} {
		fmt.Println(humanise.FileSizeIEC(size))
	}
	// Output:
	// 512 B
	// 4.1 KiB
	// 1.43 MiB
}

func ExampleMetric() {
	readings := []struct {
		value float64
		unit  string
	}{
		{1500, "V"},
		{0.005, "A"},
		{2e8, "W"},
		{math.Inf(1), "W"},
	}
	for _, r := range readings {
		formatted, err := humanise.Metric(r.value, r.unit)
		if err != nil {
			fmt.Println("error:", err)
			continue
		}
		fmt.Println(formatted)
	}
	// Output:
	// 1.5 kV
	// 5 mA
	// 200 MW
	// error: metric value must be finite, got +Inf
}

func ExampleDuration() {
	fmt.Println(humanise.Duration(3 * time.Hour))
	fmt.Println(humanise.Duration(90 * time.Second)) // one minute drops to seconds
	fmt.Println(humanise.Duration(time.Hour))        // one hour drops to minutes
	fmt.Println(humanise.Duration(36 * time.Hour))   // one day drops to hours
	// Output:
	// 3 hours
	// 90 seconds
	// 60 minutes
	// 36 hours
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

	// Someone born on 29 February ticks over on 1 March in a non-leap year.
	leapling := time.Date(2000, time.February, 29, 0, 0, 0, 0, time.UTC)
	feb28 := time.Date(2001, time.February, 28, 0, 0, 0, 0, time.UTC)
	mar1 := time.Date(2001, time.March, 1, 0, 0, 0, 0, time.UTC)
	fmt.Println(humanise.Age(leapling, feb28))
	fmt.Println(humanise.Age(leapling, mar1))
	// Output:
	// 46
	// 0
	// 1
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
	fmt.Println(humanise.Title("Taming Of The Shrew"))      // capitalised minor words are tidied
	fmt.Println(humanise.Title("the iPhone and NASA saga")) // deliberate capitals survive
	// Output:
	// Taming of the Shrew
	// Taming of the Shrew
	// The iPhone and NASA Saga
}
