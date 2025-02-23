// === Basic Command-Line Arguments ===
// 1. Print the command-line arguments passed to the program.
// 2. Retrieve and print the first argument passed to the program.
// 3. Check if at least one argument was provided, otherwise print an error.

// === Reading User Input ===
// 4. Prompt the user to enter their name and read input using `bufio.Scanner`.
// 5. Prompt the user for a password but hide the input (using `golang.org/x/term`).

// === Working with Environment Variables ===
// 6. Retrieve and print the value of an environment variable.
// 7. Set a new environment variable and retrieve it.

// === Using Flags for CLI Arguments ===
// 8. Use the `flag` package to parse a boolean flag (`-verbose`).
// 9. Parse an integer flag (`-port=8080`) and print its value.
// 10. Combine multiple flags (`-name` and `-age`) and print them.

// === Handling CLI Subcommands ===
// 11. Implement a subcommand structure (e.g., `cli-tool greet` and `cli-tool version`).
// 12. Use `flag.NewFlagSet` to handle flags for different subcommands.

// === Interactive CLI Applications ===
// 13. Create a simple interactive REPL (Read-Eval-Print Loop).
// 14. Implement a basic to-do list CLI that allows adding and listing tasks.

// === Error Handling in CLI ===
// 15. Gracefully handle missing required flags and arguments.
// 16. Provide a `--help` option to guide users.
// 17. Return exit codes (`os.Exit()`) based on success or failure.
