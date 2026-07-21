# fake

Realistic random data for tests and fixtures. Go port of the Python `fake` package.

## Usage

```go
f := fake.New(42)        // seeded, reproducible
f := fake.NewRandom()    // OS-seeded

f.FullName()             // "Rory Drake"
f.EmailFor("Jo Blogs")   // "jo.blogs@example.com"
f.Address()              // fake.Address{...}
f.Price(1, 100)          // int64 cents
f.Words(6, 12)           // lorem ipsum
dob, _ := f.RelativeTime("-30 years")
```

A `Faker` owns its own random source; the same seed always yields the same sequence. Not
safe for concurrent use - one `Faker` per goroutine.

## API

- **Numbers:** `Bool`, `Int`, `Float`, `Price`
- **Text:** `Code`, `Letters`, `Digits`, `Word`, `Words`, `Paragraph`, `Paragraphs`,
  `ParagraphsHTML`
- **People:** `FirstName`, `FirstNameMale`, `FirstNameFemale`, `LastName`, `FullName`,
  `Job`, `City`, `Suburb`, `Address`, `AddressMultiline`, `Street`, `Postcode`, `Phone`,
  `Email`, `EmailFor`, `Website`, `WebsiteFor`, `DateOfBirth`
- **Dates:** `RelativeTime` (`"now"`, `"+3 days"`, `"-40 years"`, `"2y4w7d"`), `Between`

Count arguments are `(low, high)`; pass equal values for a fixed count. Data is NZ-flavoured
and embedded at build time. `loremipsum/` is a standalone subpackage taking a `*rand.Rand`.
