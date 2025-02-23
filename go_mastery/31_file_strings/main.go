// === Reading Strings from Files ===
// 1. Open and read an entire file into a string using `os.ReadFile()`.
// 2. Open a file and read it **line by line** using `bufio.Scanner`.
// 3. Read a file **byte by byte** and print each character.

// === Writing Strings to Files ===
// 4. Write a string to a new file using `os.WriteFile()`.
// 5. Use `fmt.Fprintf()` to write formatted strings to a file.
// 6. Use `bufio.Writer` to write a large text buffer efficiently.

// === Efficiently Handling Large Files ===
// 7. Stream-read a file **chunk by chunk** instead of loading it all into memory.
// 8. Use `io.Copy()` to efficiently copy a file’s contents to another file.
// 9. Implement an **in-memory log buffer** and flush it periodically to a file.

// === Appending to Files ===
// 10. Open a file in **append mode** and write new lines without overwriting.
// 11. Write logs to a file and ensure each log entry starts on a new line.

// === Reading & Writing Compressed Files (`gzip`) ===
// 12. Read a **compressed text file** (`.gz`) and extract the contents as a string.
// 13. Write a string to a **gzip-compressed file** to save space.
// 14. Modify the compression level of a `gzip.Writer` for better performance.

// === File Metadata & Handling ===
// 15. Check if a file **exists** before opening it.
// 16. Get the **size** of a file without reading its contents.
// 17. Read the **last modified time** of a file.

// === Advanced String Processing in Files ===
// 18. Count the number of words in a large text file efficiently.
// 19. Find and print all lines in a file that contain a specific substring.
// 20. Implement a **search and replace** function that modifies a file in place.
