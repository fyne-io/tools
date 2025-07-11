package util

import "strings"

// ExtractLdflagsFromGoFlags returns the ldflags and enviroment with ldflags removed from GOFLAGS.
func ExtractLdflagsFromGoFlags(env []string) (string, []string) {
	prefix := "GOFLAGS="
	for i, v := range env {
		if strings.HasPrefix(v, prefix) {
			ldflags, goflags := extractLdFlags(strings.TrimPrefix(v, prefix))
			env[i] = goflags
			return ldflags, env
		}
	}
	return "", env
}

func extractLdFlags(goFlags string) (string, string) {
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
