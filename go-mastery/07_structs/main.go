// === Basic Struct Operations ===
// 1. Declare a struct type `Person` with fields `Name` (string) and `Age` (int).
// 2. Declare a variable of type `Person` and assign values to its fields.
// 3. Declare and initialize a struct using a struct literal.
// 4. Access and print the `Name` field of a struct variable.
// 5. Modify the `Age` field of a struct variable and print the updated struct.

// === Pointers to Structs ===
// 6. Create a pointer to a struct and modify its fields through the pointer.
// 7. Demonstrate that struct fields are accessed without explicit dereferencing.
// 8. Create a function that takes a pointer to a struct and modifies its fields.

// === Struct Methods ===
// 9. Define a method `Greet()` on `Person` that prints a greeting message.
// 10. Define a method that modifies a struct field (e.g., increment `Age` by 1).

// === Embedded Structs (Struct Composition) ===
// 11. Define a struct `Employee` that embeds a `Person` struct and has an additional `JobTitle` field.
// 12. Create an `Employee` struct and initialize its fields.
// 13. Access and print embedded struct fields (`Name`, `Age`, `JobTitle`).
// 14. Show how method promotion works in embedded structs.

// === Struct Tags (for JSON, Database, etc.) ===
// 15. Define a struct with JSON struct tags and demonstrate `json.Marshal()`.
// 16. Use `json.Unmarshal()` to convert a JSON string into a struct.

// === Working with Structs in Slices and Maps ===
// 17. Create a slice of `Person` structs and iterate over it.
// 18. Create a map where the key is a string and the value is a `Person` struct.
// 19. Modify a struct inside a slice and observe its effect.

// === Comparing Structs ===
// 20. Compare two struct instances to check if they are equal.
// 21. Show that two structs with identical fields but different field orders are still compatible.

// === Performance Considerations ===
// 22. Demonstrate why passing structs by pointer is more efficient for large structs.
// 23. Show the effect of using `omitempty` in JSON serialization.
// 24. Explain how structs are stored in memory (alignment and padding).
