// === Basic Regex Matching ===
// 1. Check if a string contains the word "Go" using `regexp.MatchString()`.
// 2. Use `regexp.Compile()` to create a reusable regex pattern.

// === Extracting Matches ===
// 3. Find the first occurrence of "Go" in a string using `FindString()`.
// 4. Extract all words that start with "G" using `FindAllString()`.

// === Using Capture Groups ===
// 5. Extract the **first word and last word** from a string using `FindStringSubmatch()`.
// 6. Use capture groups to extract the area code from a phone number `(123) 456-7890`.

// === Replacing Text ===
// 7. Replace all occurrences of "foo" with "bar" in a string using `ReplaceAllString()`.
// 8. Use regex to replace all **whitespace** with a single space.

// === Validating Input ===
// 9. Validate an **email address** using regex.
// 10. Validate a **phone number** format (e.g., `(123) 456-7890`).
// 11. Validate if a string **only contains numbers**.

// === Named Capture Groups & Advanced Matching ===
// 12. Use named capture groups to extract `firstName` and `lastName` from `"John Doe"`.
// 13. Match and extract **IPv4 addresses** from a text.
// 14. Match **dates in YYYY-MM-DD format**.
// 15. Parse a **URL** and extract the domain name.

// === Performance Considerations ===
// 16. Compare `regexp.MustCompile()` vs `regexp.Compile()` in terms of efficiency.
// 17. Benchmark regex performance on a large string.
