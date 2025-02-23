// === Basic JSON Encoding (Marshaling) ===
// 1. Define a struct `User` with fields `Name` and `Age`.
// 2. Convert a `User` struct to a JSON string and print it.
// 3. Encode a slice of `User` structs into JSON.

// === Basic JSON Decoding (Unmarshaling) ===
// 4. Convert a JSON string into a `User` struct.
// 5. Decode a JSON array into a slice of `User` structs.

// === Working with JSON Struct Tags ===
// 6. Modify the `User` struct to use custom JSON field names (`json:"full_name"`).
// 7. Make the `Age` field **optional** using the `omitempty` tag.

// === Handling Nested JSON Objects ===
// 8. Define a struct `Address` and embed it inside `User` as `HomeAddress`.
// 9. Marshal and unmarshal a `User` struct that contains an `Address` struct.

// === Handling Unknown JSON Fields ===
// 10. Parse JSON into a `map[string]interface{}` and access dynamic fields.
// 11. Use `json.RawMessage` to defer decoding of certain fields.

// === Custom JSON Marshaling & Unmarshaling ===
// 12. Implement `MarshalJSON()` for a struct to format its output differently.
// 13. Implement `UnmarshalJSON()` to customize how JSON is parsed.

// === Streaming JSON (Large Data Handling) ===
// 14. Write a large JSON object to a file using `json.Encoder`.
// 15. Read and decode JSON from a file using `json.Decoder` (streaming).

// === Error Handling in JSON Operations ===
// 16. Handle an error when decoding an invalid JSON string.
// 17. Validate required fields before unmarshaling JSON.

// === Advanced JSON Techniques ===
// 18. Convert a struct to an **indented JSON format** for pretty printing.
// 19. Encode a struct and send it over an HTTP response.
// 20. Decode JSON received from an HTTP request body.
