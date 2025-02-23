// === Basic File Operations ===
// 1. Create a new file named "example.txt" and write "Hello, Go!" to it.
// 2. Open "example.txt", read its contents, and print them to the console.
// 3. Append "This is an appended line." to "example.txt".
// 4. Delete "example.txt".

// === Reading Files in Different Ways ===
// 5. Read an entire file into a single string.
// 6. Read a file **line by line** and print each line.
// 7. Read a file **byte by byte** and print the raw bytes.
// 8. Read a file **in chunks of 512 bytes** and process it.

// === Writing to Files ===
// 9. Write multiple lines to a file using `WriteString()`.
// 10. Write data using `fmt.Fprintf()` for formatted output.
// 11. Truncate an existing file and write new content.
// 12. Write to a file using `bufio.Writer` for efficiency.

// === Working with JSON Files ===
// 13. Write a struct to a file as **JSON** using `json.Marshal()`.
// 14. Read and parse a **JSON** file into a struct.

// === Working with CSV Files ===
// 15. Write a slice of records to a **CSV** file.
// 16. Read a **CSV** file and parse the data into a struct.

// === File Permissions & Error Handling ===
// 17. Check if a file exists before opening it.
// 18. Handle errors when a file is not found.
// 19. Change the file permissions of an existing file.

// === Working with Directories ===
// 20. Create a new directory named "testdir".
// 21. List all files in a directory.
// 22. Delete a directory and its contents.

// === Advanced File Handling ===
// 23. Create and use a **temporary file**.
// 24. Create and use a **temporary directory**.
// 25. Stream large file contents without loading everything into memory.
// 26. Use **memory-mapped files** for fast access.
