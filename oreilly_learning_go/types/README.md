
# Predeclared Types

Notes on, and experiments with, predeclared types.


## Zero Value

Every declared type is given at least a zero value that varies by type.


## Booleans

Simply `true` and `false`.


## Integers

* Signed from `int8` through to `int64`, unsigned from `uint8` to `uint64`.
* The alias `byte` is provided for `uint8`.
* `int` and `uint` map to native word length.

Literals can be decimal, octal, or hexadecimal and can use underscores for
readability:

	123456
	123_456
	0xdeadbeef
	0o7875


## Floating-point

Both `float32` and `float64` are provided, defaulting to `float64`.

	123.456
	1.23e45

Complex numbers created with built-in function, or literals. They default to
`complex128` when built from a pair of `float64`, but `complex64` also exists.

	complex(20.3, 10.2)
	(19.3+36.62i)


## Runes

Single-character literal can be single quoted character, 8-bit octal, or
hexadecimal up to 32-bits. The `rune` type is an alias to `int32` (but why not
`uint32`?)

	'a'
	'\141'
	'\x61'
	'\u0061'
	'\U00000061'


## Interpreted String Literal

Zero or more runes. So-called because interpret rune literals (both numeric
and backslash escaped) into single character.


## Raw String Literal

Using backticks, no escape charaters are interpreted.

	`This is a
	muliline literal`
