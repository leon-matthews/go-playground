
# Aliases & Snippets

This is my ongoing collection of shell aliases and editor snippets for the
Go programming language.


## Bash aliases

Shell aliases for command-line quickness! Add to your *.bashrc* or equivalent.

	alias gob="go test -bench ."
	alias gof="go fmt ./..."
	alias got="go test"
	alias gov="go vet ./..."


## Editor snippets

Like shell aliases, I like to automate my text editor to emit common
boilerplate code. In the following section `%cursor%` is where your cursor
ends up after hitting *<tab>*.

	[Go]
	bench=func Benchmark%cursor%(b *testing.B) {\n\tfor i := 0; i < b.N; i++ {}\n}
	p=fmt.Println(%cursor%)
	stest=t.Run("%cursor%", func(t *testing.T) {\n})
	test=func Test%cursor%(t *testing.T) {\n}
