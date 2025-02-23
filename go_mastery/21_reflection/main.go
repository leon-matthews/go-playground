// === Basic Type Inspection ===
// 1. Print the type of a variable using `reflect.TypeOf()`.
// 2. Get the **kind** of a variable (`reflect.ValueOf(x).Kind()`) and print it.
// 3. Check if a value is a pointer using `reflect.TypeOf(x).Kind() == reflect.Ptr`.

// === Inspecting Struct Fields & Methods ===
// 4. Define a struct `Person` with fields `Name string` and `Age int`.
// 5. Use reflection to iterate over the struct fields and print their names and types.
// 6. Get and call all methods of a struct dynamically using `reflect.Type.Method()`.

// === Modifying Struct Fields Dynamically ===
// 7. Change the value of a struct field using `reflect.Value.Set()`.
// 8. Modify a private field using `reflect.Value` (if possible).
// 9. Check if a field is **exported** before modifying it.

// === Calling Functions Dynamically ===
// 10. Define a function `Add(a, b int) int` and get its `reflect.Value`.
// 11. Call `Add` dynamically using `reflect.Value.Call([]reflect.Value{})`.
// 12. Handle multiple return values when calling a function via reflection.

// === Creating Instances of Types Dynamically ===
// 13. Create a new instance of a struct using `reflect.New()`.
// 14. Dynamically assign values to a struct instance created with reflection.
// 15. Instantiate a slice dynamically using `reflect.MakeSlice()`.

// === Using Reflection for Generic-like Behavior ===
// 16. Write a function that accepts `interface{}` and prints its type dynamically.
// 17. Implement a generic-like function that prints elements of **any** slice type.
// 18. Write a function that can set a variable's value dynamically using reflection.

// === Performance Considerations & Pitfalls ===
// 19. Compare the performance of reflection vs. direct struct field access.
// 20. Demonstrate a case where reflection is **unavoidable** (e.g., JSON serialization).
// 21. Explain why excessive reflection can lead to reduced type safety and performance issues.
