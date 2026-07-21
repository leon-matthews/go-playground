package fake

import (
	"fmt"
	"math"
	"regexp"
	"strings"
	"time"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// streetEndings are the suffixes appended to a street name.
var streetEndings = []string{
	"Avenue", "Ave", "Drive", "Lane", "Place", "Pl",
	"Road", "Rd", "Street", "St", "Way",
}

// htmlEntity matches simple named HTML entities, for removal during slugging.
var htmlEntity = regexp.MustCompile(`&[a-z]+;`)

// nonSlug matches runs of characters that are not slug-safe.
var nonSlug = regexp.MustCompile(`[^a-z0-9]+`)

// accentFolder decomposes accented characters and drops their combining marks.
//
// It normalises to NFKD, removes nonspacing marks (category Mn), then recomposes
// as NFC, folding "café" to "cafe". Characters without a decomposition, such as
// 'ø' or 'ł', are left unchanged.
var accentFolder = transform.Chain(norm.NFKD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)

// City returns a random city name.
func (f *Faker) City() string {
	return f.choose(linesCities())
}

// Suburb returns a random suburb name.
func (f *Faker) Suburb() string {
	return f.choose(linesSuburbs())
}

// Job returns a random job title.
func (f *Faker) Job() string {
	return f.choose(linesJobs())
}

// FirstName returns a random first name, either male or female.
func (f *Faker) FirstName() string {
	return f.choose(linesFirstNames())
}

// FirstNameMale returns a random male first name.
func (f *Faker) FirstNameMale() string {
	return f.choose(linesFirstNamesMale())
}

// FirstNameFemale returns a random female first name.
func (f *Faker) FirstNameFemale() string {
	return f.choose(linesFirstNamesFemale())
}

// LastName returns a random last name.
func (f *Faker) LastName() string {
	return f.choose(linesLastNames())
}

// FullName returns a random first and last name.
func (f *Faker) FullName() string {
	return f.FirstName() + " " + f.LastName()
}

// Address returns a randomly assembled Address.
func (f *Faker) Address() Address {
	suburb := ""
	if f.rng.IntN(2) == 1 {
		suburb = f.Suburb()
	}
	country := ""
	if f.rng.IntN(5) == 0 {
		country = "New Zealand"
	}
	return Address{
		Address1: f.Street(),
		Suburb:   suburb,
		City:     f.City(),
		PostCode: f.Postcode(),
		Country:  country,
	}
}

// AddressMultiline returns a 2-4 line address separated by newlines.
func (f *Faker) AddressMultiline() string {
	parts := []string{f.Street()}
	if f.rng.IntN(2) == 1 {
		parts = append(parts, f.Suburb())
	}
	if f.rng.IntN(3) != 0 {
		parts = append(parts, fmt.Sprintf("%s %s", f.City(), f.Postcode()))
	} else {
		parts = append(parts, f.City())
	}
	if f.rng.IntN(5) == 0 {
		parts = append(parts, "New Zealand")
	}
	return strings.Join(parts, "\n")
}

// Street returns the street part of an address, for example "12B Happy Valley Rd".
func (f *Faker) Street() string {
	number := int(f.triangular(1, 100, 1))
	unit := ""
	if f.rng.IntN(2) == 1 {
		// Unit letter A-E, biased towards A.
		unit = string(byte('A' + int(f.triangular(0, 4, 0))))
	}
	ending := streetEndings[f.rng.IntN(len(streetEndings))]
	return fmt.Sprintf("%d%s %s %s", number, unit, f.choose(linesStreets()), ending)
}

// Postcode returns the postcode part of an address, for example "1234".
func (f *Faker) Postcode() string {
	return f.digits(4)
}

// Phone returns a random New Zealand phone number.
func (f *Faker) Phone() string {
	prefixes := []string{"021", "022", "025", "03", "04", "09"}
	prefix := prefixes[f.rng.IntN(len(prefixes))]
	separators := []string{"", "-", " "}
	separator := separators[f.rng.IntN(len(separators))]
	return strings.Join([]string{prefix, "555", f.digits(4)}, separator)
}

// Email returns a random email address with a numeric local part.
func (f *Faker) Email() string {
	return f.digits(8) + "@example.com"
}

// EmailFor returns an email address derived from the given name.
func (f *Faker) EmailFor(name string) string {
	local := strings.ReplaceAll(slug(name, 64), "-", ".")
	return local + "@example.com"
}

// Website returns a random website URL with a numeric host.
func (f *Faker) Website() string {
	return "https://" + f.digits(8) + ".com/"
}

// WebsiteFor returns a website URL derived from the given name.
func (f *Faker) WebsiteFor(name string) string {
	host := strings.ReplaceAll(slug(name, 64), "-", ".")
	return "https://" + host + ".com/"
}

// DateOfBirth returns a random date of birth for a person aged in the inclusive
// range [minAge, maxAge).
//
// Ages use the Gregorian year, so the resulting age is a half-open range. It
// panics if minAge is not less than maxAge.
func (f *Faker) DateOfBirth(minAge, maxAge int) time.Time {
	now := time.Now()
	start := now.Add(time.Duration(-float64(maxAge) * secondsPerYear * float64(time.Second)))
	end := now.Add(time.Duration(-float64(minAge) * secondsPerYear * float64(time.Second)))
	return f.Between(start, end)
}

// triangular returns a random float from a triangular distribution.
//
// It is a port of Python's random.triangular, biasing results towards mode.
func (f *Faker) triangular(low, high, mode float64) float64 {
	if high == low {
		return low
	}
	u := f.rng.Float64()
	c := (mode - low) / (high - low)
	if u > c {
		u = 1.0 - u
		c = 1.0 - c
		low, high = high, low
	}
	return low + (high-low)*math.Sqrt(u*c)
}

// slug turns arbitrary text into a hyphen-separated, URL-safe string.
//
// Accented characters are folded to their base form, so "Café René" becomes
// "cafe-rene".
func slug(text string, maxLength int) string {
	text = foldAccents(text)
	text = strings.ToLower(text)
	text = htmlEntity.ReplaceAllString(text, "")
	text = nonSlug.ReplaceAllString(text, "-")
	if maxLength > 0 && len(text) > maxLength {
		text = text[:maxLength]
	}
	return strings.Trim(text, "-")
}

// foldAccents strips diacritics from text, returning it unchanged on error.
func foldAccents(text string) string {
	folded, _, err := transform.String(accentFolder, text)
	if err != nil {
		return text
	}
	return folded
}
