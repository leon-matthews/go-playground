// Instructions:
// Complete each task below without looking at references.
// Try to recall the syntax from memory and write valid Go code.

// === Setting Up gRPC ===
// 1. Install `protoc` and the Go gRPC plugins (`protoc-gen-go` and `protoc-gen-go-grpc`).
// 2. Create a `.proto` file that defines a `Greeter` service with a `SayHello` method.
// 3. Generate Go code from the `.proto` file using `protoc`.

// === Implementing a gRPC Server ===
// 4. Implement a `GreeterServer` struct that satisfies the generated gRPC interface.
// 5. Start a gRPC server that listens on port 50051.
// 6. Modify the server to handle multiple concurrent requests using goroutines.

// === Implementing a gRPC Client ===
// 7. Write a gRPC client that connects to the `Greeter` service.
// 8. Modify the client to send multiple requests in a loop.

// === Unary RPC ===
// 9. Implement a unary RPC method that takes a request and returns a response.
// 10. Modify the unary method to return an error when the request is invalid.

// === Server Streaming RPC ===
// 11. Implement a **server streaming** RPC where the server sends multiple responses to a single request.
// 12. Modify the client to read and process the stream.

// === Client Streaming RPC ===
// 13. Implement a **client streaming** RPC where the client sends multiple requests and gets a single response.
// 14. Modify the server to aggregate client messages before responding.

// === Bi-Directional Streaming RPC ===
// 15. Implement a **bi-directional streaming** RPC where both client and server exchange messages continuously.
// 16. Modify the client to send periodic messages and process server responses.

// === Middleware with gRPC Interceptors ===
// 17. Implement a **logging interceptor** that logs all incoming gRPC requests.
// 18. Implement an **authentication interceptor** that checks metadata headers for an API key.

// === Error Handling & Deadlines ===
// 19. Implement a **gRPC deadline** on the client side and modify the server to respect it.
// 20. Modify a gRPC method to return **custom errors** using `status` and `codes` from `google.golang.org/grpc/status`.

// === gRPC With TLS ===
// 21. Set up gRPC **with TLS encryption** by generating self-signed certificates.
// 22. Modify the client to establish a **secure TLS connection** to the server.

