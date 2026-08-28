package util

import (
	"errors"
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/mod/semver"
)

var (
	rxClang   = regexp.MustCompile(`(?:(\S+)\s+)?clang\s+version\s+([0-9.]+)`)
	rxGcc     = regexp.MustCompile(`\bcc(\.exe)?\s+\(([^)]+)\)\s+([0-9.]+)`)
	rxBuiltBy = regexp.MustCompile(`built\s+by\s+(\S+)`)
)

var (
	ErrNoCompilerFound = errors.New("failed to detect compiler")
	ErrInvalidVersion  = errors.New("failed to parse version")
)

// CompilerInfo holds detected compiler information.
type CompilerInfo struct {
	Name    string
	Version string
	OS      string
}

// DetectCompiler takes the output of `cc --version` and tries to detect the
// compiler, compiler version, and os name when available.
func DetectCompiler(s, os string) (*CompilerInfo, error) {
	osname := func(s string) string {
		switch s {
		case "apple":
			return "darwin"
		case "":
			return os
		}
		return s
	}

	if m := rxClang.FindStringSubmatch(strings.ToLower(s)); len(m) == 3 {
		if !semver.IsValid("v" + m[2]) {
			return nil, ErrInvalidVersion
		}
		return &CompilerInfo{"clang", m[2], osname(m[1])}, nil
	}

	if m := rxGcc.FindStringSubmatch(strings.ToLower(s)); len(m) == 4 {
		if !semver.IsValid("v" + m[3]) {
			return nil, ErrInvalidVersion
		}

		if m[1] == ".exe" {
			if x := rxBuiltBy.FindStringSubmatch(m[2]); len(x) == 2 {
				return &CompilerInfo{"gcc", m[3], osname(x[1])}, nil
			}
		}

		if m[2] == "" || m[2] == "gcc" {
			return &CompilerInfo{"gcc", m[3], osname("")}, nil
		}

		o := append(strings.Fields(strings.Map(func(r rune) rune {
			if unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.IsSpace(r) {
				return r
			}
			return ' '
		}, m[2])), "")[0]

		return &CompilerInfo{"gcc", m[3], osname(o)}, nil
	}

	return nil, ErrNoCompilerFound
}
