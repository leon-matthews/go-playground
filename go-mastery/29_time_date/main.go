// === Getting Current Time ===
// 1. Print the **current date and time** using `time.Now()`.
// 2. Extract and print the year, month, and day from a `time.Time` value.
// 3. Extract and print the hour, minute, and second from a `time.Time` value.

// === Time Formatting & Parsing ===
// 4. Format the current time as `"2006-01-02 15:04:05"` (Go's reference time format).
// 5. Parse a string `"2023-10-15 14:30:00"` into a `time.Time` object.
// 6. Convert a `time.Time` object to an **RFC3339** formatted string.
// 7. Parse an **RFC3339** formatted string back to `time.Time`.

// === Working with Time Durations ===
// 8. Print the duration between two times using `time.Since()`.
// 9. Add **2 hours and 30 minutes** to the current time and print the result.
// 10. Subtract **15 minutes** from the current time and print the result.
// 11. Compare two time values to determine if one is **before** or **after** the other.
// 12. Round a timestamp to the nearest minute using `Truncate(time.Minute)`.

// === Working with Timezones ===
// 13. Print the current time in UTC using `time.UTC()`.
// 14. Convert the current time to New York timezone (`America/New_York`).
// 15. Convert a UTC time to **local time** and vice versa.

// === Timers & Tickers ===
// 16. Create a `time.Timer` that waits for **5 seconds** before printing a message.
// 17. Create a `time.NewTicker` that prints a message **every 2 seconds** and stops after 10 seconds.
// 18. Implement a function that **delays execution** using `time.AfterFunc()`.

// === Scheduling & Delays ===
// 19. Pause execution for **3 seconds** using `time.Sleep()`.
// 20. Use `time.After()` to **trigger an event after 2 seconds**.
// 21. Schedule a function to run **every 10 seconds** using `time.Ticker`.

// === Working with UNIX Timestamps ===
// 22. Convert the current time to a **Unix timestamp (seconds)** using `time.Unix()`.
// 23. Convert the current time to a **nanosecond Unix timestamp**.
// 24. Convert a Unix timestamp **back to a human-readable time**.

// === Handling Leap Years & Daylight Saving Time (DST) ===
// 25. Write a function to check if a given year is a **leap year**.
// 26. Determine if a given date falls within **Daylight Saving Time (DST)**.
// 27. Calculate how many days are in a given month and year (taking leap years into account).

// === High-Precision Timing ===
// 28. Measure the execution time of a function using `time.Now()` and `time.Since()`.
// 29. Use `time.AfterFunc()` to execute a function **with a delay**.
// 30. Use `time.Tick()` to implement a **basic stopwatch**.

// === Advanced Concepts ===
// 31. Calculate the **difference in days** between two dates.
// 32. Implement a countdown timer that prints time remaining every second.
// 33. Parse and format time in different **locale formats** (e.g., German, French).
// 34. Handle **time intervals longer than 24 hours** correctly.
// 35. Implement a **cron-like scheduler** that executes a task at a specific time every day.
