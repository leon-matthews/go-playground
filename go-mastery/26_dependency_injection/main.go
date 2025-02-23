// === Manual Dependency Injection (Constructor Injection) ===
// 1. Create an interface `Database` with a method `FetchData() string`.
// 2. Implement `Database` with a concrete type `SQLDatabase`.
// 3. Write a constructor function `NewService(db Database)` that injects a dependency.
// 4. Modify `NewService` to return an error if no database is provided.

// === Interface-Based Dependency Injection ===
// 5. Define a `Logger` interface with a `Log()` method.
// 6. Implement `Logger` with a `ConsoleLogger` that prints to the console.
// 7. Inject `Logger` into a struct `App` and use it in a method.

// === Injecting Dependencies into HTTP Handlers ===
// 8. Create an HTTP server with a `Handler` struct that has a `Database` dependency.
// 9. Modify the HTTP handler to fetch data from the database and return it as a response.

// === Mocking Dependencies for Testing ===
// 10. Implement a `MockDatabase` that satisfies the `Database` interface.
// 11. Write a test that injects `MockDatabase` into `NewService()` and verifies the behavior.
// 12. Create a `MockLogger` and use it in a test to check log output.

// === Using Third-Party Dependency Injection Libraries ===
// 13. Use `google/wire` to auto-generate dependency injection code.
// 14. Use `uber/fx` to create an application with dependency injection.
// 15. Compare manual DI vs. `wire` vs. `fx` in terms of maintainability.

// === Comparing DI Approaches & Best Practices ===
// 16. Explain when **manual dependency injection** is preferred over DI frameworks.
// 17. Discuss how **DI improves testability** and avoids tight coupling.
// 18. Demonstrate a case where **too much DI can overcomplicate** an application.
