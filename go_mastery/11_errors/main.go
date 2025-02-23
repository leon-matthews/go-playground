// === Basic Error Handling ===
// 1. Write a function that returns an `error` when a number is negative.
// 2. Call the function and handle the error properly.
// 3. Modify the function to return both a result and an error.

// === Creating Custom Errors ===
// 4. Define a custom error using `errors.New()`.
// 5. Define a custom error using `fmt.Errorf()`.
// 6. Use `fmt.Errorf()` with error wrapping (`%w`) to wrap an existing error.

// === Checking and Comparing Errors ===
// 7. Define a sentinel error using `var ErrSomething = errors.New("some error")`.
// 8. Check if a returned error matches the sentinel error using `errors.Is()`.
// 9. Create a wrapped error and check if it matches the original using `errors.Is()`.

// === Using `errors.As()` for Type Assertions ===
// 10. Define a custom error type using `struct` and implement the `Error()` method.
// 11. Return the custom error from a function.
// 12. Use `errors.As()` to check if an error is of the custom type.

// === Panic & Recover for Error Handling ===
// 13. Write a function that triggers a `panic` and use `recover()` to catch it.
// 14. Demonstrate why using `panic` instead of returning an error is bad practice.
// 15. Write a function that safely recovers from a `panic` and returns an error instead.

// === Best Practices for Error Handling ===
// 16. Write an example where errors are properly logged instead of ignored.
// 17. Show how to return detailed errors without exposing internal details.
// 18. Implement error handling in a function that reads from a file.
// 19. Demonstrate propagating an error up multiple function calls.
// 20. Create a wrapper function that standardizes error messages.

// === Performance Considerations ===
// 21. Explain when returning `nil` instead of an error improves efficiency.
// 22. Show an example where pre-allocating sentinel errors reduces allocations.
