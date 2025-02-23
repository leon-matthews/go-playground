// Instructions:
// Complete each task below without looking at references.
// Try to recall the syntax from memory and write valid Go code.

// === Basic Middleware ===
// 1. Write a basic middleware function that wraps an HTTP handler.
// 2. Modify the middleware to log the request method and URL.

// === Chaining Multiple Middleware Functions ===
// 3. Implement a function that chains multiple middleware functions.
// 4. Apply multiple middleware functions to a single handler.

// === Logging Middleware ===
// 5. Create a middleware that logs the request processing time.
// 6. Modify the logger to include the client IP address in the logs.

// === Authentication Middleware ===
// 7. Implement an authentication middleware that checks for an `Authorization` header.
// 8. Extend authentication middleware to verify JWT tokens.

// === Request Modification Middleware ===
// 9. Create middleware that adds a custom header to every response.
// 10. Modify request parameters inside a middleware before passing to the handler.

// === Rate Limiting Middleware ===
// 11. Implement a basic rate limiter using `time.Ticker`.
// 12. Use `golang.org/x/time/rate.Limiter` for more efficient rate limiting.

// === Recovery Middleware (Handling Panics) ===
// 13. Implement a middleware that recovers from panics and returns a 500 response.
// 14. Log panic details before recovering and continuing execution.

// === Performance Considerations & Middleware Order ===
// 15. Demonstrate how middleware execution order affects request processing.
// 16. Compare the impact of **expensive** middleware (e.g., database logging) on response time.
// 17. Discuss when to use middleware vs. writing logic directly in handlers.
