package islazy

import (
	"crypto/sha256"
	"encoding/hex"
	"unicode/utf8"
)

// Truncate shortens s to at most max bytes, keeping the start of the string.
//
// When s has to be shortened a short hash of the full value is appended, so
// that two values which only differ past the cut do not collapse onto the
// same result.
func Truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}

	sum := sha256.Sum256([]byte(s))
	suffix := "-" + hex.EncodeToString(sum[:])[:8]

	keep := max - len(suffix)
	if keep < 0 {
		keep = 0
	}

	// never slice through a multi byte rune
	for keep > 0 && !utf8.RuneStart(s[keep]) {
		keep--
	}

	return s[:keep] + suffix
}
