package reader

// cursor extracts typed values from one row's tab-separated fields.
// The first parse error is kept so a record can be built in a single declarative block.
type cursor struct {
	fields []string
	err    error
}

func (c *cursor) str(i int) string {
	return c.fields[i]
}

func (c *cursor) optionalStr(i int) string {
	return optionalString(c.fields[i])
}

func (c *cursor) list(i int) []string {
	return splitList(c.fields[i])
}

func (c *cursor) optionalInt(i int) int {
	if c.err != nil {
		return missing
	}
	n, err := optionalInt(c.fields[i])
	c.keep(err)
	return n
}

func (c *cursor) requiredInt(i int) int {
	if c.err != nil {
		return 0
	}
	n, err := requiredInt(c.fields[i])
	c.keep(err)
	return n
}

func (c *cursor) boolean(i int) bool {
	if c.err != nil {
		return false
	}
	b, err := parseBool(c.fields[i])
	c.keep(err)
	return b
}

func (c *cursor) float(i int) float64 {
	if c.err != nil {
		return 0
	}
	f, err := parseFloat(c.fields[i])
	c.keep(err)
	return f
}

func (c *cursor) characters(i int) []string {
	if c.err != nil {
		return nil
	}
	names, err := parseCharacters(c.fields[i])
	c.keep(err)
	return names
}

// keep records the first error seen.
func (c *cursor) keep(err error) {
	if c.err == nil {
		c.err = err
	}
}
