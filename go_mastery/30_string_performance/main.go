// Instructions:
// Complete each task below without looking at references.
// Try to recall the syntax from memory and write valid Go code.

// === Efficient String Concatenation ===
// 1. Concatenate two strings using the `+` operator and measure performance.
// 2. Concatenate a large number of strings using `strings.Builder` and compare efficiency.
// 3. Use `fmt.Sprintf()` for string concatenation and compare with other methods.

// === Avoiding Unnecessary String Copying ===
// 4. Modify a string efficiently without creating multiple copies.
// 5. Convert a string to a byte slice and modify it without excessive allocations.

// === Memory-Efficient Substring Extraction ===
// 6. Extract a substring without causing large memory retention issues.
// 7. Use `copy()` to create a truly independent substring in memory.

// === Preventing Large String Retention Due to Slices ===
// 8. Demonstrate how slicing a large string can keep unnecessary memory allocated.
// 9. Solve the above problem by properly creating a new independent string.

// === Handling Large Strings in Memory-Efficient Ways ===
// 10. Process a very large string without loading it entirely into memory.
// 11. Stream a string from a file efficiently using `bufio.Scanner`.
// 12. Use `bytes.Buffer` instead of `string` for large text manipulation.

// === Optimized File-Based String Processing ===
// 13. Read a large text file **line by line** without excessive memory usage.
// 14. Write a large string to a file **efficiently** using buffered writing.
// 15. Use memory-mapped files (`mmap`) to process large files efficiently.

// === Benchmarking String Operations ===
// 16. Benchmark `strings.Builder` vs. `+` operator for repeated concatenation.
// 17. Measure the performance of `fmt.Sprintf()` vs. string concatenation.
// 18. Compare performance of `strings.Split()` vs. manual string tokenization.
// 19. Benchmark `strings.Contains()` vs. `strings.Index()` for substring searches.

// === Advanced Performance Considerations ===
// 20. Use `sync.Pool` to reuse string buffers and reduce memory allocations.
// 21. Compare `strings.Clone()` vs. standard string assignment in terms of performance.
// 22. Optimize string searching using `bytes.Index()` instead of `strings.Index()`.
// 23. Use `strings.FieldsFunc()` instead of `strings.Split()` for complex splitting scenarios.
// 24. Implement a **high-performance in-memory log buffer** using `bytes.Buffer`.
