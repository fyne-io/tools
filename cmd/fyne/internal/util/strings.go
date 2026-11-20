package util

import (
	"strings"
)

// JoinSpace takes a list of strings and joins them with space as separator.
func JoinSpace(l []string) string {
	return strings.Join(l, " ")
}

// JoinComma takes a list of strings and joins them with comma as separator.
func JoinComma(l []string) string {
	return strings.Join(l, ",")
}

// SplitDot returns a slice of strings by splitting then given string by dot.
func SplitDot(s string) []string {
	return strings.Split(s, ".")
}

// SplitComma returns a slice of strings by splitting then given string by comma.
func SplitComma(s string) []string {
	return strings.Split(s, ",")
}

// SplitSlash returns a slice of strings by splitting then given string by slash.
func SplitSlash(s string) []string {
	return strings.Split(s, "/")
}
