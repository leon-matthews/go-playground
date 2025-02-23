// === Detecting Language from Text ===
// 1. Use a language detection package to identify the language of a string.
// 2. Modify the detection function to return confidence scores for multiple languages.
// 3. Identify whether a given text is written in **left-to-right (LTR)** or **right-to-left (RTL)** format.

// === Handling Different Encodings ===
// 4. Convert a string from **ISO-8859-1** to **UTF-8**.
// 5. Convert a UTF-8 encoded string to **Shift-JIS** and back.
// 6. Detect if a string contains **invalid UTF-8 characters** and fix them.

// === Transliterating Text ===
// 7. Convert a **Cyrillic text** to **Latin characters**.
// 8. Transliterate **Chinese Pinyin with tone marks** to plain Latin characters.
// 9. Convert **Arabic script** to a more readable Latin representation.

// === Sorting Unicode Strings Correctly ===
// 10. Sort a list of names that contain **accents and special characters**.
// 11. Modify the sorting to be **case-insensitive and locale-aware**.
// 12. Compare **Japanese (Kana) strings** correctly when sorting.

// === String Collation (Locale-Aware Sorting) ===
// 13. Compare two **German words** with locale-aware sorting (`ß` vs. `ss`).
// 14. Sort a list of **French words** correctly (`é` vs. `e`).
// 15. Sort a mixed list of **Chinese, Japanese, and English** words.

// === Working with Non-ASCII Characters & Normalization ===
// 16. Normalize a string from **NFD (decomposed form) to NFC (composed form)**.
// 17. Strip diacritics (accents) from characters (`é → e`).
// 18. Remove **invisible Unicode characters** from a string.
// 19. Convert **full-width** characters (`ＡＢＣ`) to **half-width** (`ABC`).
// 20. Convert Arabic digits (`١٢٣`) to standard Western digits (`123`).
