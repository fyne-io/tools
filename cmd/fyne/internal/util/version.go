package util

import (
	"strconv"
	"strings"
)

// CompareVersions orders two names by the version number in them, such as
// "33.0.2" and "36.0.0", or "android-33" and "android-36.1", so that 36
// comes after 9 where comparing the names as strings would not. The number is
// read from the first digit, one dot-separated part at a time, and stops at
// the first part that is not a number, so "36.0.0-rc1" counts as 36.0.0.
// Missing parts count as zero, so "36" and "36.0" are equal, and a name with
// no number at all counts as zero.
// It returns -1 when a is lower, 1 when a is higher, and 0 when they match.
func CompareVersions(a, b string) int {
	pa, pb := versionParts(a), versionParts(b)
	for len(pa) < len(pb) {
		pa = append(pa, 0)
	}
	for len(pb) < len(pa) {
		pb = append(pb, 0)
	}
	for i := range pa {
		if pa[i] < pb[i] {
			return -1
		}
		if pa[i] > pb[i] {
			return 1
		}
	}
	return 0
}

func versionParts(name string) []int {
	start := strings.IndexAny(name, "0123456789")
	if start < 0 {
		return nil
	}
	notDigit := func(r rune) bool { return r < '0' || r > '9' }
	var parts []int
	for _, s := range strings.Split(name[start:], ".") {
		digits := s
		if i := strings.IndexFunc(s, notDigit); i >= 0 {
			digits = s[:i]
		}
		if digits == "" {
			break
		}
		n, _ := strconv.Atoi(digits)
		parts = append(parts, n)
		if len(digits) < len(s) { // a qualifier such as "-rc1" ends the number
			break
		}
	}
	return parts
}
