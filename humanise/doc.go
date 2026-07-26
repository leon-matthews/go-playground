// Package humanise bridges the gap between how computers store values and how people read them.
//
// It loosely tracks Python's humanize library, grouping its functions into numbers,
// sizes and units, time and dates, and text.
//
// A space separates a number from its unit, following ISO 80000-1, so a file size
// reads "4.2 kB" and a voltage "1.5 kV". The deliberately compact forms are the
// exception: [WordsCompact] abbreviates to "1.2M".
//
// # Numbers
//
//   - [Comma] groups digits, eg. 1234567 becomes "1,234,567".
//   - [Underscore] groups as a Go literal, eg. 1234567 becomes "1_234_567".
//   - [Ordinal] renders an English ordinal, eg. 21 becomes "21st".
//   - [Words] names each thousandfold, eg. 1200000 becomes "1.2 million".
//   - [WordsCompact] abbreviates the same, eg. 1200000 becomes "1.2M".
//   - [Significant] rounds to significant digits, eg. 1234.567 to 3 becomes 1230.
//
// # Sizes and units
//
//   - [FileSize] renders a byte count in SI units, eg. 4200 becomes "4.2 kB".
//   - [FileSizeIEC] renders one in IEC units, eg. 4200 becomes "4.1 KiB".
//   - [Metric] renders a value with an SI prefix, eg. 1500 volts becomes "1.5 kV".
//
// # Time and dates
//
//   - [Duration] phrases a span, eg. three hours becomes "3 hours".
//   - [Relative] phrases a signed offset, eg. -5 minutes becomes "5 minutes ago".
//   - [Age] counts whole years between two dates, eg. 46.
//
// # Text
//
//   - [And] joins a list with the Oxford comma, eg. "apples, oranges, and pears".
//   - [Or] joins a list with "or", eg. "apples, oranges, or pears".
//   - [Title] capitalises a title, eg. "taming of the shrew" becomes "Taming of the Shrew".
package humanise
