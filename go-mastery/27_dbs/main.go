// === Connecting to Databases ===
// 1. Connect to a PostgreSQL database using `database/sql` and `pq` driver.
// 2. Connect to a MySQL database using the `mysql` driver.
// 3. Connect to an SQLite database using `sqlite3` driver.
// 4. Handle database connection errors properly.

// === Executing Queries ===
// 5. Write a `SELECT` query to retrieve all users from a `users` table.
// 6. Execute an `INSERT` query to add a new user to the database.
// 7. Write an `UPDATE` query to modify a user's email address.
// 8. Execute a `DELETE` query to remove a user by ID.

// === Using `database/sql` vs. ORMs ===
// 9. Implement the `SELECT` query using raw SQL (`database/sql`).
// 10. Implement the `SELECT` query using an ORM (`gorm`).
// 11. Compare raw SQL vs. ORM in terms of performance and maintainability.

// === Transactions & Rollbacks ===
// 12. Implement a database transaction to transfer money between two accounts.
// 13. Rollback the transaction if any query fails.
// 14. Modify the transaction to use `defer tx.Rollback()` for safety.

// === Prepared Statements & SQL Injection Prevention ===
// 15. Rewrite the `INSERT` query using a prepared statement.
// 16. Modify a `SELECT` query to use query parameters and prevent SQL injection.

// === Connection Pooling & Performance Optimization ===
// 17. Configure **connection pooling** using `sql.DB.SetMaxOpenConns()`.
// 18. Set **idle connections** using `sql.DB.SetMaxIdleConns()`.
// 19. Benchmark the performance of database queries with different connection pool settings.

// === Handling Migrations ===
// 20. Use `go-migrate` to apply a schema migration for a new table.
// 21. Implement a function that **auto-runs pending migrations** on application startup.
