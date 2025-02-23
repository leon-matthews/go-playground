// === Basic HTTP Server ===
// 1. Create an HTTP server that listens on port 8080 and responds with "Hello, World!".
// 2. Modify the server to handle multiple endpoints (`/`, `/about`, `/contact`).

// === Handling Different HTTP Methods ===
// 3. Create a handler that responds differently based on `GET` and `POST` requests.
// 4. Implement a handler for `PUT` and `DELETE` requests.

// === Working with Query Parameters & URL Variables ===
// 5. Extract a query parameter from a `GET` request (`/search?query=golang`).
// 6. Use URL path variables to capture dynamic parts (`/users/{id}`).

// === Handling JSON Requests & Responses ===
// 7. Return a JSON response with `application/json` content-type.
// 8. Accept a JSON request, parse it into a struct, and return a response.

// === Middleware in HTTP Servers ===
// 9. Implement a **logging middleware** that logs each request.
// 10. Implement an **authentication middleware** that restricts access.

// === Graceful Shutdown & Timeouts ===
// 11. Use `context.WithTimeout()` to set a timeout for HTTP requests.
// 12. Implement **graceful shutdown** for the HTTP server.

// === Serving Static Files ===
// 13. Serve static files from a `public/` directory.
// 14. Serve an HTML page when visiting `/home`.

// === Advanced Topics ===
// 15. Implement WebSockets to support real-time communication.
// 16. Implement an HTTP server that supports **CORS (Cross-Origin Resource Sharing)**.
