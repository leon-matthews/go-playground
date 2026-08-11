# Let's Go

by Alex Edwards.

## Password hashing

The book hashes passwords using `bcrypt` rather than the author's own wrapper
around `argon2`, which I'd rather use. I'm going to 'trust the process' and 
follow along with the text for now. For reference:

- The package `github.com/alexedwards/argon2id` creates a 'standard' PHC string format.
- bcrypt needs only CHAR(60) for its output, argon2 PHC needs VARCHAR(255)
- argon2 can get very expensive, both in CPU and memory, so we may want to inject 
  its config into the users model,  rather than having it as a global, for testing.
- bcrypt has a max input length of 72 bytes, which has been built into this site.

## TODO

Continue from Chapter 10.4: "User login"
