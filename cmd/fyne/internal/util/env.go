package util

import "strings"

// ExtractLdflagsFromGoFlags returns the ldflags and environment with ldflags removed from GOFLAGS.
func ExtractLdflagsFromGoFlags(env []string) (string, []string) {
	prefix := "GOFLAGS="
	for i, v := range env {
		if strings.HasPrefix(v, prefix) {
			ldflags, goflags := ExtractLdFlags(strings.TrimPrefix(v, prefix))
			env[i] = prefix + goflags
			return ldflags, env
		}
	}
	return "", env
}

// ExtractLdFlags extracts ldflags from the value of GOFLAGS environment variable, returns ldflags and new GOFLAGS.
func ExtractLdFlags(goFlags string) (string, string) {
	if goFlags == "" {
		return "", ""
	}

	flags := strings.Fields(goFlags)
	ldflags := ""
	newGoFlags := ""

	for _, flag := range flags {
		if strings.HasPrefix(flag, "-ldflags=") {
			ldflags += strings.TrimPrefix(flag, "-ldflags=") + " "
		} else {
			newGoFlags += flag + " "
		}
	}

	ldflags = strings.TrimSpace(ldflags)
	newGoFlags = strings.TrimSpace(newGoFlags)

	return ldflags, newGoFlags
}
