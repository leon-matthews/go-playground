package writer

import (
	"flag"
	"fmt"
	"os"
)

const permissions = 0o600
const Usage = `Usage: writefile -size SIZE_BYTES PATH

Creates the file PATH, containing SIZE_BYTES bytes, all zero.

Example: writefile -size 1000 zeroes.dat`

func WriteToFile(path string, data []byte) error {
	err := os.WriteFile(path, data, permissions)
	if err != nil {
		return err
	}

	// os.WriteFile doesn't change permissions of pre-existing file
	return os.Chmod(path, permissions)
}

func Main() {
	if len(os.Args) < 2 {
		fmt.Println(Usage)
		return
	}
	size := flag.Int("size", 0, "Size in bytes")
	flag.Parse()
	if len(flag.Args()) < 1 {
		fmt.Fprintln(os.Stderr, Usage)
		os.Exit(1)
	}
	err := WriteToFile(flag.Args()[0], make([]byte, *size))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
