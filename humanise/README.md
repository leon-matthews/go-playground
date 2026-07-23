# humanise

Package `humanise` formats values for human consumption. It loosely tracks Python's
[humanize](https://github.com/python-humanize/humanize) library, grouping its functions into
numbers, sizes and units, time and dates, and text.

## Numbers

| Function       | Example call               | Result        |
| -------------- | -------------------------- | ------------- |
| `Comma`        | `Comma(1234567)`           | `1,234,567`   |
| `Underscore`   | `Underscore(1234567)`      | `1_234_567`   |
| `Ordinal`      | `Ordinal(21)`              | `21st`        |
| `Words`        | `Words(1200000)`           | `1.2 million` |
| `WordsCompact` | `WordsCompact(1200000)`    | `1.2M`        |
| `Significant`  | `Significant(1234.567, 3)` | `1230`        |

## Sizes & units

| Function      | Example call        | Result   |
| ------------- | ------------------- | -------- |
| `FileSize`    | `FileSize(4200)`    | `4.2kB`  |
| `FileSizeIEC` | `FileSizeIEC(4200)` | `4.1KiB` |
| `Metric`      | `Metric(1500, "V")` | `1.5 kV` |

## Time & dates

| Function   | Example call                 | Result          |
| ---------- | ---------------------------- | --------------- |
| `Duration` | `Duration(3 * time.Hour)`    | `3 hours`       |
| `Relative` | `Relative(-5 * time.Minute)` | `5 minutes ago` |
| `Age`      | `Age(born, today)`           | `46`            |

## Text

| Function | Example call                   | Result                |
| -------- | ------------------------------ | --------------------- |
| `And`    | `And([]string{"a", "b", "c"})` | `a, b, and c`         |
| `Or`     | `Or([]string{"a", "b", "c"})`  | `a, b, or c`          |
| `Title`  | `Title("taming of the shrew")` | `Taming of the Shrew` |
