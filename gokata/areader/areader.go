// Package areader contains a custom type called A that satisfies the io.Reader
// interface. The A's Read method reads len(p) bytes into p. Each byte that is
// read represents the ASCII character 'A'.
package areader

// A is a bare type that happens to implements [io.Reader].
// type Reader interface {
//     Read(p []byte) (n int, err error)
// }
type A struct{}

// Read reads up to len(p) bytes into p, but every byte is 'A'
// Returns number of bytes written and any error encountered.
func (A) Read(p []byte) (int, error) {
	for i := range p {
		p[i] = 'A'
	}
	return len(p), nil
}
