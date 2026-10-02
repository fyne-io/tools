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

// safeShellChars are the characters that never need quoting in a shell command.
const safeShellChars = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
	"0123456789" +
	"_+-./:=,@%"

// ShellQuote quotes s for use in a shell command if it contains characters that
// would otherwise be interpreted, following the style of commands like ls.
func ShellQuote(s string) string {
	needsQuote := s == ""
	for _, r := range s {
		if !strings.ContainsRune(safeShellChars, r) {
			needsQuote = true
			break
		}
	}
	if !needsQuote {
		return s
	}

	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

// ShellQuoteCommand returns the name and arguments of a command as a single
// string, quoting each part that needs it so that the result can be run in a
// shell.
func ShellQuoteCommand(name string, args []string) string {
	quoted := make([]string, 0, len(args)+1)
	quoted = append(quoted, ShellQuote(name))
	for _, arg := range args {
		quoted = append(quoted, ShellQuote(arg))
	}

	return JoinSpace(quoted)
}
