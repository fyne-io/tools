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
