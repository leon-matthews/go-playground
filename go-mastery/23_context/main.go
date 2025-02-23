// === Basic Context Usage ===
// 1. Create a root context using `context.Background()`.
// 2. Use `context.TODO()` in a function where context will be added later.

// === Using `context.WithCancel` ===
// 3. Create a `context.WithCancel()` and pass it to a goroutine.
// 4. Cancel the context after 2 seconds and observe the goroutine exit.
// 5. Modify the goroutine to **listen for context cancellation** using `<-ctx.Done()`.

// === Using `context.WithTimeout` ===
// 6. Create a `context.WithTimeout()` that times out after 3 seconds.
// 7. Check if `ctx.Err()` is `context.DeadlineExceeded` after the timeout.
// 8. Use `select` to wait for either a result **or** the timeout.

// === Using `context.WithDeadline` ===
// 9. Create a `context.WithDeadline()` that expires at a specific time.
// 10. Modify a function to accept `context.Context` and respect the deadline.

// === Passing Context Through Functions ===
// 11. Write a function `fetchData(ctx context.Context)` that listens for cancellation.
// 12. Call `fetchData()` with a timeout-based context and observe behavior.

// === Handling Context Cancellation Properly ===
// 13. Use `defer cancel()` immediately after creating a `context.WithCancel()`.
// 14. Ensure **all spawned goroutines exit** when a parent context is canceled.

// === Using Context for Database Queries & API Calls ===
// 15. Simulate an **API request with a timeout** using `context.WithTimeout()`.
// 16. Modify a **database query function** to accept `context.Context`.
// 17. Implement **graceful shutdown** by canceling all active contexts on exit.

// === Context Best Practices & Performance Considerations ===
// 18. Explain why **passing context.Context as a function parameter** is better than storing it globally.
// 19. Compare `context.WithCancel()` vs. `context.WithTimeout()` and when to use each.
// 20. Discuss **memory leaks caused by unhandled context cancellations** and how to avoid them.
