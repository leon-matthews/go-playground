package mytypes

type MyInt int

func (x MyInt) Double() {
	x *= MyInt(2)
}
