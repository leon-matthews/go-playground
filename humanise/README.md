# humanise

Package `humanise` formats numbers for human consumption: thousands separators, file
sizes, durations, ordinals, SI prefixes, and large-number words. It loosely tracks
Python's [humanize](https://github.com/python-humanize/humanize) library.

## Functions

| Function       | Example call                | Result          |
| -------------- | --------------------------- | --------------- |
| `Comma`        | `Comma(1234567)`            | `1,234,567`     |
| `Underscore`   | `Underscore(1234567)`       | `1_234_567`     |
| `Words`        | `Words(1200000)`            | `1.2 million`   |
| `WordsCompact` | `WordsCompact(1200000)`     | `1.2M`          |
| `Ordinal`      | `Ordinal(21)`               | `21st`          |
| `FileSize`     | `FileSize(4200)`            | `4.2kB`         |
| `FileSizeIEC`  | `FileSizeIEC(4200)`         | `4.1KiB`        |
| `Metric`       | `Metric(1500, "V")`         | `1.5 kV`        |
| `Duration`     | `Duration(3 * time.Hour)`   | `3 hours`       |
| `Relative`     | `Relative(-5 * time.Minute)`| `5 minutes ago` |
| `Significant`  | `Significant(1234.567, 3)`  | `1230`          |
