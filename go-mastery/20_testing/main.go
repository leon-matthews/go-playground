// === Basic Assertions ===
// 1. Write a test that checks if `2 + 2` equals `4`.
// 2. Test if the function `Square(n int) int` correctly returns `n * n`.

// === Table-Driven Tests ===
// 3. Implement a table-driven test for a function that doubles a number.
// 4. Use `t.Run()` inside the test loop to name sub-tests.

// === Error Handling in Tests ===
// 5. Test if a function correctly returns an error when given invalid input.
// 6. Use `t.Fatalf()` to fail a test if an unexpected error occurs.

// === Benchmarking Performance ===
// 7. Benchmark a function that performs string concatenation.
// 8. Benchmark an optimized version of the function and compare results.

// === Mocking Dependencies ===
// 9. Create an interface `Database` with a method `FetchData() string`.
// 10. Implement a real struct `SQLDatabase` and a mock `MockDatabase`.
// 11. Write a test that verifies function behavior using the mock.

// === Testing HTTP Handlers ===
// 12. Create an HTTP handler that responds with "Hello, World!".
// 13. Write a test that makes an HTTP request to the handler and checks the response.

// === Concurrency Testing ===
// 14. Test a function that spawns multiple goroutines to ensure correct synchronization.
// 15. Use `t.Parallel()` inside a table-driven test to run sub-tests concurrently.

// === Using Temporary Files in Tests ===
// 16. Write a test that creates a temporary file, writes to it, and verifies the content.

// === Code Coverage ===
// 17. Run `go test -cover` and identify untested code paths.
