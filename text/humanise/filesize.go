package humanise

import (
	"fmt"
	"math"
	"strconv"
)

// Unit suffixes for SI (base-1000) and IEC (base-1024) multiples, from kilo up.
var (
	siSuffixes  = []string{"kB", "MB", "GB", "TB", "PB", "EB", "ZB", "YB", "RB", "QB"}
	iecSuffixes = []string{"KiB", "MiB", "GiB", "TiB", "PiB", "EiB", "ZiB", "YiB", "RiB", "QiB"}
)

// FileSize renders a byte count as a short, human-readable string using SI units.
//
// Sizes that round below 1000 show as a whole byte count; larger sizes use decimal
// multiples of 1000 (kB, MB, GB …) and up to three significant figures, so
// 4200 becomes "4.2kB". It returns an error for negative or non-finite sizes.
func FileSize(size float64) (string, error) {
	return formatFileSize(size, 1000, siSuffixes)
}

// FileSizeIEC renders a byte count as a short, human-readable string using IEC units.
//
// Sizes that round below 1024 show as a whole byte count; larger sizes use binary
// multiples of 1024 (KiB, MiB, GiB …) and up to three significant figures, so
// 4200 becomes "4.1KiB". It returns an error for negative or non-finite sizes.
func FileSizeIEC(size float64) (string, error) {
	return formatFileSize(size, 1024, iecSuffixes)
}

func formatFileSize(size, base float64, suffixes []string) (string, error) {
	switch {
	case math.IsNaN(size) || math.IsInf(size, 0):
		return "", fmt.Errorf("file size must be finite, got %v", size)
	case size < 0:
		return "", fmt.Errorf("file size must not be negative, got %v", size)
	case size == 0:
		return "0B", nil
	}

	// Compare the rounded count so that sizes like 999.7 become "1kB", not "1000B".
	if bytes := math.RoundToEven(size); bytes < base {
		return strconv.FormatFloat(bytes, 'f', 0, 64) + "B", nil
	}

	// Divide down until the value fits its unit, or units run out.
	index := 0
	size /= base
	for size >= base && index < len(suffixes)-1 {
		size /= base
		index++
	}
	rounded, err := Significant(size, 3)
	if err != nil {
		return "", err
	}
	return strconv.FormatFloat(rounded, 'g', -1, 64) + suffixes[index], nil
}
