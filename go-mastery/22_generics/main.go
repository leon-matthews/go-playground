// === Basic Generic Functions ===
// 1. Write a generic function that returns the larger of two values.
// 2. Implement a generic function that swaps two values.
// 3. Write a generic function that reverses a slice of any type.

// === Generic Structs & Methods ===
// 4. Define a generic struct `Box[T]` that holds a single value of any type.
// 5. Implement a method `Get()` for `Box[T]` that returns its value.
// 6. Implement a method `Set()` that updates the value inside `Box[T]`.

// === Type Constraints in Generics ===
// 7. Write a generic function that **only works with numbers** using `constraints.Ordered`.
// 8. Write a function that **only works with comparable types** (`comparable`).
// 9. Implement a generic **Min()** and **Max()** function for slices of ordered types.

// === Using `any` for Maximum Flexibility ===
// 10. Write a function that accepts **any** type (`any`) and prints its value.
// 11. Modify the function to use reflection (`reflect.TypeOf`) to print its actual type.

// === Generic Data Structures ===
// 12. Implement a **Stack[T]** data structure with `Push()` and `Pop()` methods.
// 13. Implement a **Queue[T]** with `Enqueue()` and `Dequeue()` methods.
// 14. Create a **LinkedList[T]** that supports insertion and traversal.

// === Combining Generics with Reflection ===
// 15. Write a function that **uses reflection** to inspect a generic type at runtime.
// 16. Implement a function that takes a slice of **any** type and prints the element types dynamically.

// === Performance Considerations & When to Use Generics ===
// 17. Compare the performance of generics vs. interface-based implementations.
// 18. Explain when using **generics is better than using `interface{}`**.
// 19. Implement a benchmark test that compares generics with reflection-based solutions.
