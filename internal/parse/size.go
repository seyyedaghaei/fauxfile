package parse

import (
	"errors"
	"strconv"
	"strings"
	"unicode"
)

// Bytes parses a size string into bytes.
// Supports: raw bytes (e.g. "1024"), units b, k/kb, m/mb, g/gb (case insensitive).
// Decimals allowed (e.g. "1.5mb"). Leading/trailing space is trimmed.
func Bytes(s string) (int64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, errors.New("empty size")
	}
	s = strings.ToLower(s)

	// Find where the number ends (digit or decimal point).
	i := 0
	for i < len(s) && (unicode.IsDigit(rune(s[i])) || s[i] == '.') {
		i++
	}
	if i == 0 {
		return 0, errors.New("invalid size: no number")
	}
	numStr := s[:i]
	unit := strings.TrimSpace(s[i:])

	n, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return 0, err
	}
	if n < 0 {
		return 0, errors.New("size must be non-negative")
	}

	var mult float64 = 1
	switch unit {
	case "", "b", "byte", "bytes":
		mult = 1
	case "k", "kb":
		mult = 1024
	case "m", "mb":
		mult = 1024 * 1024
	case "g", "gb":
		mult = 1024 * 1024 * 1024
	default:
		return 0, errors.New("unknown unit: " + unit)
	}

	out := n * mult
	if out > 1e18 {
		return 0, errors.New("size too large")
	}
	return int64(out), nil
}
