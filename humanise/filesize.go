package humanise

import "strconv"

// Unit suffixes for SI (base-1000) and IEC (base-1024) multiples, from kilo up.
var (
	siSuffixes  = []string{"kB", "MB", "GB", "TB", "PB", "EB", "ZB", "YB", "RB", "QB"}
	iecSuffixes = []string{"KiB", "MiB", "GiB", "TiB", "PiB", "EiB", "ZiB", "YiB", "RiB", "QiB"}
)

// FileSize renders a byte count as a short, human-readable string using SI units.
//
// Counts below 1000 render as a whole number of bytes; larger counts use decimal
// multiples of 1000 (kB, MB, GB …) carrying up to three significant figures, so 4200
// becomes "4.2kB". A negative count renders by its magnitude with a leading minus.
func FileSize(size int64) string {
	return formatFileSize(size, 1000, siSuffixes)
}

// FileSizeIEC renders a byte count as a short, human-readable string using IEC units.
//
// Counts below 1024 render as a whole number of bytes; larger counts use binary
// multiples of 1024 (KiB, MiB, GiB …) carrying up to three significant figures, so 4200
// becomes "4.1KiB". A negative count renders by its magnitude with a leading minus.
func FileSizeIEC(size int64) string {
	return formatFileSize(size, 1024, iecSuffixes)
}

// formatFileSize renders size in multiples of base, labelled with suffixes from kilo up.
func formatFileSize(size int64, base float64, suffixes []string) string {
	// magnitude handles MinInt64 without overflow; the sign is restored afterwards.
	bytes := magnitude(size)
	sign := ""
	if size < 0 {
		sign = "-"
	}

	if bytes < base {
		return sign + strconv.FormatFloat(bytes, 'f', 0, 64) + "B"
	}

	mantissa, index := scale(bytes, base, len(suffixes)-1)
	return sign + formatMantissa(mantissa) + suffixes[index]
}
