// === Basic Interface Usage ===
// 1. Define an interface `Speaker` with a method `Speak() string`.
// 2. Implement `Speaker` with a struct `Person` that returns "Hello!" from `Speak()`.
// 3. Implement `Speaker` with another struct `Dog` that returns "Woof!" from `Speak()`.
// 4. Write a function that accepts a `Speaker` interface and calls `Speak()`.

// === Interface Composition ===
// 5. Define an interface `Writer` with a method `Write() string`.
// 6. Define an interface `Reader` with a method `Read() string`.
// 7. Define an interface `ReadWriter` that embeds `Reader` and `Writer`.
// 8. Implement `ReadWriter` with a struct `File`.

// === Empty Interfaces (`interface{}`) and Type Assertions ===
// 9. Write a function that takes an `interface{}` and prints its value.
// 10. Use type assertion to check if a variable of type `interface{}` is a `string`.
// 11. Modify the function to handle different types (`int`, `bool`, `float64`).

// === Type Switching with Interfaces ===
// 12. Write a function that uses a **type switch** to handle different types inside an `interface{}`.
// 13. Extend the function to handle `string`, `int`, `float64`, and `bool`.

// === Interfaces in Structs ===
// 14. Define an interface `Notifier` with a method `Notify() string`.
// 15. Create a struct `EmailNotifier` that implements `Notifier`.
// 16. Create a struct `SMSNotifier` that implements `Notifier`.
// 17. Store multiple `Notifier` instances in a slice and iterate over them.

// === Using Interfaces for Dependency Injection ===
// 18. Define an interface `Logger` with a method `Log(message string)`.
// 19. Implement `Logger` with a struct `ConsoleLogger` that prints messages to stdout.
// 20. Implement `Logger` with a struct `FileLogger` that writes messages to a file.
// 21. Write a function that accepts a `Logger` and logs a message.

// === Mocking with Interfaces ===
// 22. Define an interface `Database` with a method `FetchData() string`.
// 23. Implement `Database` with a real struct `SQLDatabase`.
// 24. Implement `Database` with a mock struct `MockDatabase` for testing.
// 25. Write a function that depends on `Database` and demonstrate switching implementations.

// === Performance Considerations ===
// 26. Compare performance between using an interface and a concrete struct.
// 27. Show how interface values require dynamic type resolution and its effect on performance.
// 28. Explain when **not** to use interfaces in Go.
